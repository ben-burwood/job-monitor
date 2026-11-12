import type { JobState } from "@/types/job";

// daisyUI colour classes per state, written out in full so Tailwind can detect them
export const STATE_CLASSES: Record<JobState, { badge: string; dot: string }> = {
  success: { badge: "badge-success", dot: "bg-success" },
  failure: { badge: "badge-error", dot: "bg-error" },
  error: { badge: "badge-error", dot: "bg-error" },
  warning: { badge: "badge-warning", dot: "bg-warning" },
  running: { badge: "badge-info", dot: "bg-info animate-pulse" },
  cancelled: { badge: "badge-neutral", dot: "bg-neutral" },
  skipped: { badge: "badge-ghost", dot: "bg-base-300" },
  neutral: { badge: "badge-ghost", dot: "bg-base-300" },
  none: { badge: "badge-ghost", dot: "bg-base-300" },
};
