<template>
    <div class="card bg-base-100 shadow-sm">
        <div class="card-body p-0">
            <h2 class="card-title px-5 pt-4">{{ title }}</h2>
            <div class="overflow-x-auto">
                <table class="table">
                    <thead>
                        <tr>
                            <th v-if="source === 'github'">Repo</th>
                            <th>{{ source === "github" ? "Workflow" : "Job" }}</th>
                            <th>Status</th>
                            <th>Last run</th>
                            <th>Runtime</th>
                            <th></th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="job in jobs" :key="job.id" class="hover:bg-base-200">
                            <td v-if="source === 'github'" class="font-medium">
                                {{ job.repo }}
                                <span v-if="job.branch" class="badge badge-ghost badge-sm ms-1">{{ job.branch }}</span>
                            </td>
                            <td class="whitespace-nowrap" :class="{ 'font-medium': source === 'glue' }">{{ job.name }}</td>
                            <td>
                                <StatusBadge :job="job" />
                                <div v-if="job.error" class="text-xs text-error mt-1 max-w-md break-words">{{ job.error }}</div>
                            </td>
                            <td class="whitespace-nowrap">
                                <template v-if="job.startedAt">
                                    {{ formatTimestamp(new Date(job.startedAt)) }}
                                    <span class="opacity-60">· {{ formatRelative(new Date(job.startedAt), now) }}</span>
                                </template>
                                <span v-else class="opacity-50">—</span>
                            </td>
                            <td class="whitespace-nowrap">
                                <span v-if="job.startedAt">{{ formatDuration(runtime(job, now)) }}</span>
                                <span v-else class="opacity-50">—</span>
                            </td>
                            <td>
                                <a v-if="job.url" :href="job.url" target="_blank" rel="noopener" class="link link-primary">Open</a>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import StatusBadge from "@/components/StatusBadge.vue";
import type { JobSource, JobStatus } from "@/types/job";
import { formatDuration, formatRelative, formatTimestamp } from "@/utils/format";

defineProps<{
    title: string;
    source: JobSource;
    jobs: JobStatus[];
    now: Date;
}>();

// Running jobs keep ticking from their start time; finished jobs use the server-reported duration
function runtime(job: JobStatus, now: Date): number {
    if (job.state !== "running" || !job.startedAt) return job.durationSeconds;
    return Math.max(0, Math.floor((now.getTime() - new Date(job.startedAt).getTime()) / 1000));
}
</script>
