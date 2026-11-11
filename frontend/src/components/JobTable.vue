<template>
    <div class="card bg-base-100 shadow-sm">
        <div class="card-body p-0">
            <h2 class="card-title px-5 pt-4">{{ title }}</h2>
            <div class="overflow-x-auto">
                <table class="table">
                    <thead>
                        <tr>
                            <th class="w-0"></th>
                            <th v-if="source === 'github'">Repo</th>
                            <th>{{ source === "github" ? "Workflow" : "Job" }}</th>
                            <th>Status</th>
                            <th>Last run</th>
                            <th>Runtime</th>
                            <th></th>
                        </tr>
                    </thead>
                    <tbody>
                        <template v-for="job in jobs" :key="job.id">
                            <tr
                                class="hover:bg-base-200 cursor-pointer"
                                :class="{ 'bg-base-200': expanded.has(job.id) }"
                                tabindex="0"
                                :aria-expanded="expanded.has(job.id)"
                                title="Show run history"
                                @click="toggle(job.id)"
                                @keydown.enter.prevent="toggle(job.id)"
                                @keydown.space.prevent="toggle(job.id)"
                            >
                                <td class="pe-0">
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg"
                                        width="16"
                                        height="16"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        stroke-width="2"
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                        class="opacity-60 transition-transform"
                                        :class="{ 'rotate-90': expanded.has(job.id) }"
                                    >
                                        <path d="m9 18 6-6-6-6" />
                                    </svg>
                                </td>
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
                                    <a v-if="job.url" :href="job.url" target="_blank" rel="noopener" class="link link-primary" @click.stop>Open</a>
                                </td>
                            </tr>
                            <tr v-if="expanded.has(job.id)">
                                <td :colspan="source === 'github' ? 7 : 6" class="bg-base-200 p-0">
                                    <JobHistory :job-id="job.id" :now="now" :refresh-token="refreshToken" />
                                </td>
                            </tr>
                        </template>
                    </tbody>
                </table>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import JobHistory from "@/components/JobHistory.vue";
import StatusBadge from "@/components/StatusBadge.vue";
import type { JobSource, JobStatus } from "@/types/job";
import { formatDuration, formatRelative, formatTimestamp, runtime } from "@/utils/format";
import { reactive } from "vue";

defineProps<{
    title: string;
    source: JobSource;
    jobs: JobStatus[];
    now: Date;
    refreshToken: number; // bumped on force refresh so open histories refetch
}>();

const expanded = reactive(new Set<string>());

function toggle(id: string) {
    if (expanded.has(id)) expanded.delete(id);
    else expanded.add(id);
}
</script>
