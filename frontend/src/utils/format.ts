import type { JobRun } from "@/types/job";

export function formatRelative(ts: Date, now: Date): string {
  const secs = Math.max(0, Math.floor((now.getTime() - ts.getTime()) / 1000));
  if (secs < 60) return `${secs}s ago`;
  if (secs < 3600) return `${Math.floor(secs / 60)}m ago`;
  if (secs < 86400) return `${Math.floor(secs / 3600)}h ago`;
  return `${Math.floor(secs / 86400)}d ago`;
}

export function formatTimestamp(ts: Date): string {
  const date = ts.toLocaleDateString("en-GB", { day: "numeric", month: "short", year: "numeric" });
  const time = ts.toLocaleTimeString("en-GB", { hour: "2-digit", minute: "2-digit" });
  return `${date} · ${time}`;
}

export function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
  return `${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`;
}

// Running jobs keep ticking from their start time; finished jobs use the server-reported duration
export function runtime(run: JobRun, now: Date): number {
  if (run.state !== "running" || !run.startedAt) return run.durationSeconds;
  return Math.max(0, Math.floor((now.getTime() - new Date(run.startedAt).getTime()) / 1000));
}
