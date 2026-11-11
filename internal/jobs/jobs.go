package jobs

import (
	"context"
	"time"

	"job-monitor/internal/config"
	"job-monitor/internal/github"
	"job-monitor/internal/glue"
)

type Source string

const (
	SourceGithub Source = "github"
	SourceGlue   Source = "glue"
)

// State is the normalised status of a job's latest run, shared across sources
type State string

const (
	StateSuccess   State = "success"
	StateFailure   State = "failure"
	StateCancelled State = "cancelled"
	StateSkipped   State = "skipped"
	StateWarning   State = "warning"
	StateNeutral   State = "neutral"
	StateRunning   State = "running"
	StateNoRuns    State = "none"
	StateError     State = "error" // the status could not be fetched
)

// HistoryWindow is how far back a job's run history reaches
const HistoryWindow = 7 * 24 * time.Hour

// Run is a single run of a job, normalised across sources
type Run struct {
	State    State      `json:"state"`
	Label    string     `json:"label"`
	Number   int        `json:"number,omitempty"`  // GitHub run number
	Title    string     `json:"title,omitempty"`   // GitHub run title, usually the head commit message
	Trigger  string     `json:"trigger,omitempty"` // what started the run, e.g. push, schedule or a Glue trigger name
	Message  string     `json:"message,omitempty"` // failure reason reported by the source (Glue)
	URL      string     `json:"url,omitempty"`
	Started  *time.Time `json:"startedAt,omitempty"`
	Duration int64      `json:"durationSeconds"` // as of FetchedAt; for running jobs the UI extends this from startedAt
}

var noRuns = Run{State: StateNoRuns, Label: "No runs"}

// JobStatus is the latest run of a single tracked job
type JobStatus struct {
	ID     string `json:"id"`
	Source Source `json:"source"`
	Index  int    `json:"index"` // position within its source in the config
	Name   string `json:"name"`  // workflow filename or Glue job name
	Repo   string `json:"repo,omitempty"`
	Branch string `json:"branch,omitempty"`
	Run
	Error   string    `json:"error,omitempty"`
	Fetched time.Time `json:"fetchedAt"`
}

// History is a job's runs within HistoryWindow, newest first
type History struct {
	ID      string    `json:"id"`
	Since   time.Time `json:"since"`
	Runs    []Run     `json:"runs"`
	Fetched time.Time `json:"fetchedAt"`
}

// source fetches the runs of a single configured job
type source interface {
	// latest fills in status with the job's most recent run
	latest(ctx context.Context, status *JobStatus) error
	// history returns the job's runs started at or after since, newest first
	history(ctx context.Context, since time.Time) ([]Run, error)
}

// errSource stands in for every job of a source whose client could not be created (e.g. GITHUB_TOKEN unset)
type errSource struct{ err error }

func (s errSource) latest(context.Context, *JobStatus) error          { return s.err }
func (s errSource) history(context.Context, time.Time) ([]Run, error) { return nil, s.err }

// orErr returns src, or an errSource if its client failed to be created
func orErr(src source, err error) source {
	if err != nil {
		return errSource{err}
	}
	return src
}

// Job is a configured job
type Job struct {
	ID   string
	base JobStatus // identifying fields, copied into each fetched status
	src  source
}

// Service fetches and caches job statuses and histories
type Service struct {
	jobs      []Job
	statuses  ttlCache[*JobStatus]
	histories ttlCache[*History]
}

// NewService builds the configured jobs. Clients are created eagerly; a missing client
// is reported as an error on each of that source's jobs.
func NewService(ctx context.Context, cfg *config.Config, ttl time.Duration) *Service {
	s := &Service{}
	s.statuses.ttl = ttl
	s.histories.ttl = ttl

	if len(cfg.Github) > 0 {
		gh, err := github.NewClientFromEnv()
		for i, j := range cfg.Github {
			s.add(JobStatus{ID: "github:" + j.Repo + "/" + j.Workflow, Source: SourceGithub, Index: i, Name: j.Workflow, Repo: j.Repo},
				orErr(&githubSource{gh: gh, repo: j.Repo, workflow: j.Workflow}, err))
		}
	}

	if len(cfg.Glue) > 0 {
		gl, err := glue.NewClientFromEnv(ctx)
		for i, j := range cfg.Glue {
			s.add(JobStatus{ID: "glue:" + j.Name, Source: SourceGlue, Index: i, Name: j.Name},
				orErr(&glueSource{gl: gl, name: j.Name}, err))
		}
	}

	return s
}

func (s *Service) add(base JobStatus, src source) {
	s.jobs = append(s.jobs, Job{ID: base.ID, base: base, src: src})
}

// Jobs returns all configured jobs in config order
func (s *Service) Jobs() []Job {
	return s.jobs
}

// Job returns the configured job with the given ID
func (s *Service) Job(id string) (Job, bool) {
	for _, job := range s.jobs {
		if job.ID == id {
			return job, true
		}
	}
	return Job{}, false
}

// Status returns the job's latest status, served from cache unless expired or force is set.
// Fetch errors are reported (and cached) as an error status.
func (s *Service) Status(ctx context.Context, job Job, force bool) *JobStatus {
	status, _ := s.statuses.get(ctx, job.ID, force, func(ctx context.Context) (*JobStatus, error) {
		status := job.base
		if err := job.src.latest(ctx, &status); err != nil {
			status.State, status.Label, status.Error = StateError, "Error", err.Error()
		}
		status.Fetched = time.Now().UTC()
		return &status, nil
	})
	return status
}

// History returns the job's runs within HistoryWindow, served from cache unless expired or force is set.
// Fetch errors are returned rather than cached.
func (s *Service) History(ctx context.Context, job Job, force bool) (*History, error) {
	return s.histories.get(ctx, job.ID, force, func(ctx context.Context) (*History, error) {
		since := time.Now().UTC().Add(-HistoryWindow)
		runs, err := job.src.history(ctx, since)
		if err != nil {
			return nil, err
		}
		if runs == nil {
			runs = []Run{} // encode as [] rather than null
		}
		return &History{ID: job.ID, Since: since, Runs: runs, Fetched: time.Now().UTC()}, nil
	})
}
