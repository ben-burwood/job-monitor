package jobs

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue/types"

	"job-monitor/internal/github"
)

var t0 = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

func TestApplyGithubRun(t *testing.T) {
	cases := []struct {
		name         string
		run          github.WorkflowRun
		wantState    State
		wantLabel    string
		wantDuration int64
	}{
		{"success", github.WorkflowRun{Status: "completed", Conclusion: "success", RunStartedAt: t0, UpdatedAt: t0.Add(90 * time.Second)}, StateSuccess, "Success", 90},
		{"timed out", github.WorkflowRun{Status: "completed", Conclusion: "timed_out", RunStartedAt: t0, UpdatedAt: t0.Add(time.Hour)}, StateFailure, "Timed out", 3600},
		{"unknown conclusion", github.WorkflowRun{Status: "completed", Conclusion: "stale", RunStartedAt: t0, UpdatedAt: t0}, StateNeutral, "Stale", 0},
		{"in progress", github.WorkflowRun{Status: "in_progress", RunStartedAt: t0, UpdatedAt: t0}, StateRunning, "In progress", 300},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var s Run
			applyGithubRun(&s, &c.run, t0.Add(5*time.Minute))
			if s.State != c.wantState || s.Label != c.wantLabel || s.Duration != c.wantDuration {
				t.Errorf("got (%s, %q, %d), want (%s, %q, %d)", s.State, s.Label, s.Duration, c.wantState, c.wantLabel, c.wantDuration)
			}
		})
	}
}

func TestApplyGlueRun(t *testing.T) {
	cases := []struct {
		name         string
		run          types.JobRun
		wantState    State
		wantLabel    string
		wantDuration int64
	}{
		{"succeeded", types.JobRun{JobRunState: types.JobRunStateSucceeded, StartedOn: aws.Time(t0), ExecutionTime: 42}, StateSuccess, "Succeeded", 42},
		{"stopped", types.JobRun{JobRunState: types.JobRunStateStopped, StartedOn: aws.Time(t0), ExecutionTime: 7}, StateCancelled, "Stopped", 7},
		{"running", types.JobRun{JobRunState: types.JobRunStateRunning, StartedOn: aws.Time(t0)}, StateRunning, "Running", 300},
		{"waiting without start", types.JobRun{JobRunState: types.JobRunStateWaiting}, StateRunning, "Waiting", 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var s Run
			applyGlueRun(&s, &c.run, t0.Add(5*time.Minute))
			if s.State != c.wantState || s.Label != c.wantLabel || s.Duration != c.wantDuration {
				t.Errorf("got (%s, %q, %d), want (%s, %q, %d)", s.State, s.Label, s.Duration, c.wantState, c.wantLabel, c.wantDuration)
			}
		})
	}
}
