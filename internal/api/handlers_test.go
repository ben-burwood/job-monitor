package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"job-monitor/internal/config"
	"job-monitor/internal/github"
	"job-monitor/internal/jobs"
)

func newTestService(t *testing.T) *jobs.Service {
	t.Setenv(github.TokenEnvVar, "") // force a deterministic, offline error status
	cfg := &config.Config{Github: []config.GithubJob{{Repo: "a/b", Workflow: "one.yml"}, {Repo: "a/b", Workflow: "two.yml"}}}
	return jobs.NewService(context.Background(), cfg, time.Minute)
}

func TestJobsHandlerSSE(t *testing.T) {
	rw := httptest.NewRecorder()
	JobsHandlerSSE(newTestService(t))(rw, httptest.NewRequest("GET", "/api/jobs-sse", nil))

	body := rw.Body.String()
	if !strings.HasSuffix(body, "event: done\ndata: {}\n\n") {
		t.Errorf("expected trailing done event, got %q", body)
	}

	names := map[string]bool{}
	for _, line := range strings.Split(body, "\n") {
		data, ok := strings.CutPrefix(line, "data: {\"id\"")
		if !ok {
			continue
		}
		var status jobs.JobStatus
		if err := json.Unmarshal([]byte("{\"id\""+data), &status); err != nil {
			t.Fatalf("failed to decode event %q: %v", line, err)
		}
		if status.State != jobs.StateError || status.Error != github.ErrTokenNotSet.Error() {
			t.Errorf("expected token error, got %+v", status)
		}
		names[status.Name] = true
	}
	if len(names) != 2 || !names["one.yml"] || !names["two.yml"] {
		t.Errorf("expected one event per job, got %v", names)
	}
}
func TestJobHistoryHandler(t *testing.T) {
	cases := []struct {
		name     string
		url      string
		wantCode int
		wantBody string
	}{
		{"unknown job", "/api/jobs/history?id=github:a/b/missing.yml", http.StatusNotFound, "unknown job"},
		{"fetch error", "/api/jobs/history?id=github:a/b/one.yml", http.StatusBadGateway, github.ErrTokenNotSet.Error()},
		{"fetch error with days", "/api/jobs/history?id=github:a/b/one.yml&days=28", http.StatusBadGateway, github.ErrTokenNotSet.Error()},
		{"unsupported days", "/api/jobs/history?id=github:a/b/one.yml&days=3", http.StatusBadRequest, "days must be one of"},
		{"invalid days", "/api/jobs/history?id=github:a/b/one.yml&days=abc", http.StatusBadRequest, "days must be one of"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rw := httptest.NewRecorder()
			JobHistoryHandler(newTestService(t))(rw, httptest.NewRequest("GET", c.url, nil))
			if rw.Code != c.wantCode || !strings.Contains(rw.Body.String(), c.wantBody) {
				t.Errorf("got %d %q, want %d containing %q", rw.Code, rw.Body.String(), c.wantCode, c.wantBody)
			}
		})
	}
}
