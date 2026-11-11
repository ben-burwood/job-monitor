export type JobSource = "github" | "glue";

export type JobState = "success" | "failure" | "cancelled" | "skipped" | "warning" | "neutral" | "running" | "none" | "error";

export interface JobRun {
  state: JobState;
  label: string;
  number?: number;
  title?: string;
  trigger?: string;
  message?: string;
  url?: string;
  startedAt?: string;
  durationSeconds: number;
}

export interface JobStatus extends JobRun {
  id: string;
  source: JobSource;
  index: number;
  name: string;
  repo?: string;
  branch?: string;
  error?: string;
  fetchedAt: string;
}

export interface JobHistory {
  id: string;
  since: string;
  runs: JobRun[];
  fetchedAt: string;
}
