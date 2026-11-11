<template>
    <span class="badge whitespace-nowrap" :class="badgeClass" :title="job.error">
        <span v-if="job.state === 'running'" class="loading loading-spinner loading-xs"></span>
        {{ job.label }}
    </span>
</template>

<script setup lang="ts">
import type { JobState, JobStatus } from "@/types/job";
import { computed } from "vue";

const props = defineProps<{
    job: JobStatus;
}>();

const BADGE_CLASSES: Record<JobState, string> = {
    success: "badge-success",
    failure: "badge-error",
    error: "badge-warning",
    warning: "badge-warning",
    running: "badge-info",
    cancelled: "badge-neutral",
    skipped: "badge-ghost",
    neutral: "badge-ghost",
    none: "badge-ghost",
};

const badgeClass = computed(() => BADGE_CLASSES[props.job.state] ?? "badge-ghost");
</script>
