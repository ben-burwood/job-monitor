package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"sync"

	"job-monitor/internal/jobs"
)

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

// JobHistoryHandler handles requests for a single job's run history (?id=<job id>&days=<n>), newest first
func JobHistoryHandler(svc *jobs.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		job, ok := svc.Job(r.URL.Query().Get("id"))
		if !ok {
			http.Error(w, "unknown job", http.StatusNotFound)
			return
		}
		days := jobs.DefaultHistoryDays
		if d := r.URL.Query().Get("days"); d != "" {
			n, err := strconv.Atoi(d)
			if err != nil || !slices.Contains(jobs.HistoryDays, n) {
				http.Error(w, fmt.Sprintf("days must be one of %v", jobs.HistoryDays), http.StatusBadRequest)
				return
			}
			days = n
		}
		force := r.URL.Query().Get("refresh") == "true"

		history, err := svc.History(r.Context(), job, days, force)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(history)
	}
}
