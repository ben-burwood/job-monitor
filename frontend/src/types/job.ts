export type JobSource = "github" | "glue";

export type JobState = "success" | "failure" | "cancelled" | "skipped" | "warning" | "neutral" | "running" | "none" | "error";

export interface JobStatus {
  id: string;
  source: JobSource;
  index: number;
  name: string;
  repo?: string;
  branch?: string;
  state: JobState;
  label: string;
  error?: string;
  url?: string;
  startedAt?: string;
  durationSeconds: number;
  fetchedAt: string;
}
