package jobs

import (
	"context"
	"strings"
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

func fetchGithub(ctx context.Context, gh *github.Client, status JobStatus) *JobStatus {
	branch, err := gh.GetDefaultBranch(ctx, status.Repo)
	if err != nil {
		return status.withError(err)
	}
	status.Branch = branch

	run, err := gh.GetLatestRun(ctx, status.Repo, status.Name, branch)
	if err != nil {
		return status.withError(err)
	}
	if run == nil {
		return status.withNoRuns()
	}
	applyGithubRun(&status, run, time.Now())
	return &status
}

func applyGithubRun(status *JobStatus, run *github.WorkflowRun, now time.Time) {
	started := run.RunStartedAt
	status.Started = &started
	status.URL = run.HTMLURL

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

func fetchGlue(ctx context.Context, gl *glue.Client, status JobStatus) *JobStatus {
	run, err := gl.GetLatestRun(ctx, status.Name)
	if err != nil {
		return status.withError(err)
	}
	if run == nil {
		return status.withNoRuns()
	}
	status.URL = gl.ConsoleURL(status.Name, aws.ToString(run.Id))
	applyGlueRun(&status, run, time.Now())
	return &status
}

func applyGlueRun(status *JobStatus, run *types.JobRun, now time.Time) {
	status.Started = run.StartedOn

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
