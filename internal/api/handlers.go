package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"job-monitor/internal/jobs"
)

// AllJobsHandler handles requests for the status of all jobs, in config order
func AllJobsHandler(svc *jobs.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		force := r.URL.Query().Get("refresh") == "true"

		all := svc.Jobs()
		results := make([]*jobs.JobStatus, len(all))
		var wg sync.WaitGroup
		for i, job := range all {
			wg.Go(func() {
				results[i] = svc.Status(r.Context(), job, force)
			})
		}
		wg.Wait()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}

// JobHistoryHandler handles requests for a single job's run history (?id=<job id>), newest first
func JobHistoryHandler(svc *jobs.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		job, ok := svc.Job(r.URL.Query().Get("id"))
		if !ok {
			http.Error(w, "unknown job", http.StatusNotFound)
			return
		}
		force := r.URL.Query().Get("refresh") == "true"

		history, err := svc.History(r.Context(), job, force)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(history)
	}
}

// JobsHandlerSSE handles Server Sent Events (SSE) Endpoint, sending each job's status as it resolves.
func JobsHandlerSSE(svc *jobs.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		force := r.URL.Query().Get("refresh") == "true"

		var mu sync.Mutex
		var wg sync.WaitGroup
		for _, job := range svc.Jobs() {
			wg.Go(func() {
				status := svc.Status(r.Context(), job, force)
				b, _ := json.Marshal(status)
				mu.Lock()
				fmt.Fprintf(w, "data: %s\n\n", b) // Send Single Job Status
				w.(http.Flusher).Flush()
				mu.Unlock()
			})
		}
		wg.Wait()

		fmt.Fprintf(w, "event: done\ndata: {}\n\n") // Send Done Event
		w.(http.Flusher).Flush()
	}
}
