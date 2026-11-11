package jobs

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue/types"

	"job-monitor/internal/github"
	"job-monitor/internal/glue"
)

var githubConclusions = map[string]struct {
	state State
	label string
}{
	"success":         {StateSuccess, "Success"},
	"failure":         {StateFailure, "Failure"},
	"timed_out":       {StateFailure, "Timed out"},
	"startup_failure": {StateFailure, "Startup failure"},
	"cancelled":       {StateCancelled, "Cancelled"},
	"skipped":         {StateSkipped, "Skipped"},
	"action_required": {StateWarning, "Action required"},
	"neutral":         {StateNeutral, "Neutral"},
}

var glueTerminalStates = map[types.JobRunState]struct {
	state State
	label string
}{
	types.JobRunStateSucceeded: {StateSuccess, "Succeeded"},
	types.JobRunStateFailed:    {StateFailure, "Failed"},
	types.JobRunStateTimeout:   {StateFailure, "Timeout"},
	types.JobRunStateError:     {StateFailure, "Error"},
	types.JobRunStateExpired:   {StateFailure, "Expired"},
	types.JobRunStateStopped:   {StateCancelled, "Stopped"},
}

// branchTTL is how long a repo's default branch is cached before being looked up again
const branchTTL = time.Hour

// githubSource is a GitHub Actions workflow, tracked on its repo's default branch
type githubSource struct {
	gh       *github.Client
	repo     string
	workflow string

	mu              sync.Mutex
	branch          string
	branchExpiresAt time.Time
}

// defaultBranch returns the repo's default branch, looked up at most once per branchTTL
func (s *githubSource) defaultBranch(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if time.Now().Before(s.branchExpiresAt) {
		return s.branch, nil
	}
	branch, err := s.gh.GetDefaultBranch(ctx, s.repo)
	if err != nil {
		return "", err
	}
	s.branch, s.branchExpiresAt = branch, time.Now().Add(branchTTL)
	return branch, nil
}

func (s *githubSource) latest(ctx context.Context, status *JobStatus) error {
	branch, err := s.defaultBranch(ctx)
	if err != nil {
		return err
	}
	status.Branch = branch

	run, err := s.gh.GetLatestRun(ctx, s.repo, s.workflow, branch)
	if err != nil {
		return err
	}
	if run == nil {
		status.Run = noRuns
		return nil
	}
	applyGithubRun(&status.Run, run, time.Now())
	return nil
}

func (s *githubSource) history(ctx context.Context, since time.Time) ([]Run, error) {
	branch, err := s.defaultBranch(ctx)
	if err != nil {
		return nil, err
	}
	runs, err := s.gh.ListRuns(ctx, s.repo, s.workflow, branch, since)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	history := make([]Run, len(runs))
	for i := range runs {
		applyGithubRun(&history[i], &runs[i], now)
	}
	return history, nil
}

func applyGithubRun(status *Run, run *github.WorkflowRun, now time.Time) {
	started := run.RunStartedAt
	status.Started = &started
	status.URL = run.HTMLURL
	status.Number = run.RunNumber
	status.Title = run.DisplayTitle
	status.Trigger = run.Event

	if run.Status != "completed" {
		status.State = StateRunning
		status.Label = titleCase(run.Status)
		status.Duration = secondsBetween(started, now)
		return
	}

	if c, ok := githubConclusions[run.Conclusion]; ok {
		status.State, status.Label = c.state, c.label
	} else {
		status.State, status.Label = StateNeutral, titleCase(run.Conclusion)
	}
	status.Duration = secondsBetween(started, run.UpdatedAt)
}

// glueSource is an AWS Glue job
type glueSource struct {
	gl   *glue.Client
	name string
}

func (s *glueSource) latest(ctx context.Context, status *JobStatus) error {
	run, err := s.gl.GetLatestRun(ctx, s.name)
	if err != nil {
		return err
	}
	if run == nil {
		status.Run = noRuns
		return nil
	}
	status.URL = s.gl.ConsoleURL(s.name, aws.ToString(run.Id))
	applyGlueRun(&status.Run, run, time.Now())
	return nil
}

func (s *glueSource) history(ctx context.Context, since time.Time) ([]Run, error) {
	runs, err := s.gl.ListRuns(ctx, s.name, since)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	history := make([]Run, len(runs))
	for i := range runs {
		history[i].URL = s.gl.ConsoleURL(s.name, aws.ToString(runs[i].Id))
		applyGlueRun(&history[i], &runs[i], now)
	}
	return history, nil
}

func applyGlueRun(status *Run, run *types.JobRun, now time.Time) {
	status.Started = run.StartedOn
	status.Trigger = aws.ToString(run.TriggerName)
	status.Message = aws.ToString(run.ErrorMessage)

	if s, ok := glueTerminalStates[run.JobRunState]; ok {
		status.State, status.Label = s.state, s.label
		status.Duration = int64(run.ExecutionTime)
		return
	}

	status.State = StateRunning
	status.Label = titleCase(string(run.JobRunState))
	if run.StartedOn != nil {
		status.Duration = secondsBetween(*run.StartedOn, now)
	}
}

func secondsBetween(from, to time.Time) int64 {
	return max(0, int64(to.Sub(from).Seconds()))
}

// titleCase converts API states such as "in_progress" or "RUNNING" to "In progress" / "Running"
func titleCase(s string) string {
	if s == "" {
		return "Unknown"
	}
	s = strings.ToLower(strings.ReplaceAll(s, "_", " "))
	return strings.ToUpper(s[:1]) + s[1:]
}
