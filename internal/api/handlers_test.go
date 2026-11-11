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

func TestAllJobsHandler(t *testing.T) {
	rw := httptest.NewRecorder()
	AllJobsHandler(newTestService(t))(rw, httptest.NewRequest("GET", "/api/jobs", nil))

	if rw.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rw.Code)
	}
	var statuses []jobs.JobStatus
	if err := json.NewDecoder(rw.Body).Decode(&statuses); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(statuses) != 2 || statuses[0].Name != "one.yml" || statuses[1].Name != "two.yml" {
		t.Fatalf("expected both jobs in config order, got %+v", statuses)
	}
	if statuses[0].State != jobs.StateError || statuses[0].Error != github.ErrTokenNotSet.Error() {
		t.Errorf("expected token error, got %+v", statuses[0])
	}
}

func TestJobsHandlerSSE(t *testing.T) {
	rw := httptest.NewRecorder()
	JobsHandlerSSE(newTestService(t))(rw, httptest.NewRequest("GET", "/api/jobs-sse", nil))

	body := rw.Body.String()
	if n := strings.Count(body, "data: {\"id\""); n != 2 {
		t.Errorf("expected 2 job events, got %d in %q", n, body)
	}
	if !strings.HasSuffix(body, "event: done\ndata: {}\n\n") {
		t.Errorf("expected trailing done event, got %q", body)
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
