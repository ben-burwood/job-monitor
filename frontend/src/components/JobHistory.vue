<template>
    <div class="px-5 py-4 flex flex-col gap-4">
        <div v-if="loading && !history" class="text-center py-4">
            <span class="loading loading-dots loading-md"></span>
        </div>
        <div v-else-if="error" class="text-sm text-error break-words">Failed to load run history: {{ error }}</div>
        <template v-else-if="history">
            <!-- SUMMARY -->
            <div class="flex flex-wrap items-baseline gap-x-8 gap-y-2 text-sm">
                <div>
                    <span class="opacity-60">Runs (7d)</span> <span class="font-semibold">{{ history.runs.length }}</span>
                </div>
                <div>
                    <span class="opacity-60">Success rate</span>
                    <span class="font-semibold">{{ stats.successRate === null ? "—" : `${stats.successRate}%` }}</span>
                </div>
                <div>
                    <span class="opacity-60">Failures</span>
                    <span class="font-semibold" :class="{ 'text-error': stats.failures > 0 }">{{ stats.failures }}</span>
                </div>
                <div>
                    <span class="opacity-60">Avg runtime</span>
                    <span class="font-semibold">{{ stats.avgDuration === null ? "—" : formatDuration(stats.avgDuration) }}</span>
                </div>
                <span v-if="loading" class="loading loading-spinner loading-xs"></span>
            </div>

            <!-- PER-DAY STRIP (oldest → newest) -->
            <div class="grid gap-1" :style="{ gridTemplateColumns: `repeat(${days.length}, minmax(0, 1fr))` }">
                <div v-for="day in days" :key="day.key" class="flex flex-col gap-1 min-w-0">
                    <div class="text-xs opacity-60 whitespace-nowrap">{{ day.label }}</div>
                    <div class="flex flex-wrap gap-0.5 min-h-3">
                        <component
                            :is="run.url ? 'a' : 'span'"
                            v-for="(run, i) in day.runs"
                            :key="i"
                            :href="run.url"
                            target="_blank"
                            rel="noopener"
                            class="size-3 rounded-sm"
                            :class="DOT_CLASSES[run.state] ?? 'bg-base-300'"
                            :title="dotTitle(run)"
                        />
                    </div>
                </div>
            </div>

            <!-- RUN LIST -->
            <div v-if="history.runs.length" class="max-h-80 overflow-y-auto rounded-box bg-base-100">
                <table class="table table-sm">
                    <thead class="sticky top-0 bg-base-100">
                        <tr>
                            <th>Status</th>
                            <th>Started</th>
                            <th>Runtime</th>
                            <th>Details</th>
                            <th></th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="(run, i) in history.runs" :key="i">
                            <td><StatusBadge :job="run" /></td>
                            <td class="whitespace-nowrap">
                                <template v-if="run.startedAt">
                                    {{ formatTimestamp(new Date(run.startedAt)) }}
                                    <span class="opacity-60">· {{ formatRelative(new Date(run.startedAt), now) }}</span>
                                </template>
                                <span v-else class="opacity-50">—</span>
                            </td>
                            <td class="whitespace-nowrap">
                                <span v-if="run.startedAt">{{ formatDuration(runtime(run, now)) }}</span>
                                <span v-else class="opacity-50">—</span>
                            </td>
                            <td class="max-w-md">
                                <div class="truncate">
                                    <span v-if="run.number" class="opacity-60">#{{ run.number }}</span>
                                    {{ run.title }}
                                    <span v-if="run.trigger" class="badge badge-ghost badge-sm">{{ run.trigger }}</span>
                                </div>
                                <div v-if="run.message" class="text-xs text-error break-words">{{ run.message }}</div>
                            </td>
                            <td>
                                <a v-if="run.url" :href="run.url" target="_blank" rel="noopener" class="link link-primary">Open</a>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
            <div v-else class="text-sm italic opacity-60">No runs in the last 7 days.</div>
        </template>
    </div>
</template>

<script setup lang="ts">
import StatusBadge from "@/components/StatusBadge.vue";
import { SERVER_URL } from "@/main";
import type { JobHistory, JobRun, JobState } from "@/types/job";
import { formatDuration, formatRelative, formatTimestamp, runtime } from "@/utils/format";
import { computed, onMounted, ref, watch } from "vue";

const props = defineProps<{
    jobId: string;
    now: Date;
    refreshToken: number;
}>();

const DOT_CLASSES: Record<JobState, string> = {
    success: "bg-success",
    failure: "bg-error",
    error: "bg-warning",
    warning: "bg-warning",
    running: "bg-info animate-pulse",
    cancelled: "bg-neutral",
    skipped: "bg-base-300",
    neutral: "bg-base-300",
    none: "bg-base-300",
};

const history = ref<JobHistory | null>(null);
const loading = ref(false);
const error = ref<string | null>(null);

async function load(force = false) {
    loading.value = true;
    error.value = null;
    try {
        const resp = await fetch(`${SERVER_URL}/jobs/history?id=${encodeURIComponent(props.jobId)}${force ? "&refresh=true" : ""}`);
        if (!resp.ok) throw new Error((await resp.text()).trim() || `HTTP ${resp.status}`);
        history.value = await resp.json();
    } catch (e) {
        error.value = e instanceof Error ? e.message : String(e);
    } finally {
        loading.value = false;
    }
}

onMounted(() => load());
watch(
    () => props.refreshToken,
    () => load(true),
);

const stats = computed(() => {
    const runs = history.value?.runs ?? [];
    const successes = runs.filter((r) => r.state === "success").length;
    const failures = runs.filter((r) => r.state === "failure").length;
    const finished = runs.filter((r) => r.state !== "running" && r.startedAt);
    return {
        failures,
        successRate: successes + failures ? Math.round((successes / (successes + failures)) * 100) : null,
        avgDuration: finished.length ? Math.round(finished.reduce((sum, r) => sum + r.durationSeconds, 0) / finished.length) : null,
    };
});

// One column per local calendar day from the start of the history window to today, runs oldest first
const days = computed(() => {
    if (!history.value) return [];
    const dayKey = (d: Date) => `${d.getFullYear()}-${d.getMonth()}-${d.getDate()}`;

    const buckets: { key: string; label: string; runs: JobRun[] }[] = [];
    const cursor = new Date(history.value.since);
    cursor.setHours(0, 0, 0, 0);
    while (cursor <= props.now) {
        buckets.push({ key: dayKey(cursor), label: cursor.toLocaleDateString("en-GB", { weekday: "short", day: "numeric" }), runs: [] });
        cursor.setDate(cursor.getDate() + 1);
    }

    const byKey = new Map(buckets.map((b) => [b.key, b]));
    for (const run of [...history.value.runs].reverse()) {
        if (run.startedAt) byKey.get(dayKey(new Date(run.startedAt)))?.runs.push(run);
    }
    return buckets;
});

function dotTitle(run: JobRun): string {
    const parts = [run.label];
    if (run.startedAt) parts.push(formatTimestamp(new Date(run.startedAt)));
    if (run.startedAt) parts.push(formatDuration(runtime(run, props.now)));
    if (run.title) parts.push(run.title);
    return parts.join(" · ");
}
</script>
