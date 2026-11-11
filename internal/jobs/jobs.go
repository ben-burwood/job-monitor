package jobs

import (
	"context"
	"sync"
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

// JobStatus is the latest run of a single tracked job
type JobStatus struct {
	ID       string     `json:"id"`
	Source   Source     `json:"source"`
	Index    int        `json:"index"` // position within its source in the config
	Name     string     `json:"name"`  // workflow filename or Glue job name
	Repo     string     `json:"repo,omitempty"`
	Branch   string     `json:"branch,omitempty"`
	State    State      `json:"state"`
	Label    string     `json:"label"`
	Error    string     `json:"error,omitempty"`
	URL      string     `json:"url,omitempty"`
	Started  *time.Time `json:"startedAt,omitempty"`
	Duration int64      `json:"durationSeconds"` // as of FetchedAt; for running jobs the UI extends this from startedAt
	Fetched  time.Time  `json:"fetchedAt"`
}

// Job is a configured job that can fetch its own status
type Job struct {
	ID    string
	fetch func(ctx context.Context) *JobStatus
}

// Service fetches and caches job statuses
type Service struct {
	jobs  []Job
	ttl   time.Duration
	cache sync.Map // map[string]cacheEntry
}

type cacheEntry struct {
	status    *JobStatus
	expiresAt time.Time
}

// NewService builds the configured jobs. Clients are created eagerly; a missing client
// (e.g. GITHUB_TOKEN unset) is reported as an error on each of that source's jobs.
func NewService(ctx context.Context, cfg *config.Config, ttl time.Duration) *Service {
	s := &Service{ttl: ttl}

	if len(cfg.Github) > 0 {
		gh, ghErr := github.NewClientFromEnv()
		for i, j := range cfg.Github {
			base := JobStatus{ID: "github:" + j.Repo + "/" + j.Workflow, Source: SourceGithub, Index: i, Name: j.Workflow, Repo: j.Repo}
			s.jobs = append(s.jobs, Job{ID: base.ID, fetch: func(ctx context.Context) *JobStatus {
				status := base
				if ghErr != nil {
					return status.withError(ghErr)
				}
				return fetchGithub(ctx, gh, status)
			}})
		}
	}

	if len(cfg.Glue) > 0 {
		gl, glErr := glue.NewClientFromEnv(ctx)
		for i, j := range cfg.Glue {
			base := JobStatus{ID: "glue:" + j.Name, Source: SourceGlue, Index: i, Name: j.Name}
			s.jobs = append(s.jobs, Job{ID: base.ID, fetch: func(ctx context.Context) *JobStatus {
				status := base
				if glErr != nil {
					return status.withError(glErr)
				}
				return fetchGlue(ctx, gl, status)
			}})
		}
	}

	return s
}

// Jobs returns all configured jobs in config order
func (s *Service) Jobs() []Job {
	return s.jobs
}

// Status returns the job's latest status, served from cache unless expired or force is set
func (s *Service) Status(ctx context.Context, job Job, force bool) *JobStatus {
	if !force {
		if v, ok := s.cache.Load(job.ID); ok {
			entry := v.(cacheEntry)
			if time.Now().Before(entry.expiresAt) {
				return entry.status
			}
		}
	}

	status := job.fetch(ctx)
	status.Fetched = time.Now().UTC()
	s.cache.Store(job.ID, cacheEntry{status: status, expiresAt: time.Now().Add(s.ttl)})
	return status
}

func (s JobStatus) withError(err error) *JobStatus {
	s.State = StateError
	s.Label = "Error"
	s.Error = err.Error()
	return &s
}

func (s JobStatus) withNoRuns() *JobStatus {
	s.State = StateNoRuns
	s.Label = "No runs"
	return &s
}
