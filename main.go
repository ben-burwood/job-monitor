package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"job-monitor/internal/api"
	"job-monitor/internal/config"
	"job-monitor/internal/jobs"
)

const (
	ConfigPathEnvVar = "JOB_MONITOR_CONFIG_PATH"

	cacheTTL = 60 * time.Second
)

func main() {
	cfg, err := config.LoadConfig(os.Getenv(ConfigPathEnvVar))
	if err != nil {
		log.Fatal(err)
	}

	svc := jobs.NewService(context.Background(), cfg, cacheTTL)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/jobs-sse", api.JobsHandlerSSE(svc))
	mux.HandleFunc("GET /api/jobs/history", api.JobHistoryHandler(svc))

	// Serve Static Frontend
	mux.Handle("/", http.FileServer(http.Dir("./frontend/dist")))

	log.Fatal(http.ListenAndServe("[::]:8080", api.CORSMiddleware(mux)))
}
