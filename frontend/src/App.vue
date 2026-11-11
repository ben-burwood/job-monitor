<template>
    <div class="min-h-screen w-full flex flex-col">
        <!-- HEADER -->
        <div class="navbar bg-base-100 shadow-sm">
            <div class="flex items-center gap-4 lg:ml-20">
                <img src="/logo.svg" alt="Job Dashboard Logo" class="h-10" />
                <span class="text-2xl font-bold">Job Dashboard</span>
            </div>
            <div class="flex-1"></div>
            <button
                class="btn btn-ghost btn-circle lg:mr-20"
                :disabled="loading"
                title="Force refresh (bypass cache)"
                aria-label="Force refresh"
                @click="loadJobs(true)"
            >
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="24"
                    height="24"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    :class="{ 'animate-spin': loading }"
                >
                    <path d="M21 12a9 9 0 1 1-2.64-6.36" />
                    <path d="M21 3v6h-6" />
                </svg>
            </button>
        </div>

        <!-- BODY -->
        <div class="p-5 flex-1 bg-base-200">
            <div class="max-w-6xl mx-auto flex flex-col gap-6" v-if="jobs.length">
                <JobTable v-if="githubJobs.length" title="GitHub Actions" source="github" :jobs="githubJobs" :now="now" />
                <JobTable v-if="glueJobs.length" title="AWS Glue" source="glue" :jobs="glueJobs" :now="now" />
            </div>
            <div v-else-if="loading" class="text-center mt-10">
                <span class="loading loading-dots loading-xl"></span>
            </div>
            <div v-else-if="failed" class="text-center text-error">Failed to load jobs.</div>
            <div v-else class="text-center italic">No jobs.</div>
        </div>

        <!-- FOOTER -->
        <footer class="p-4 bg-base-300 flex items-center justify-end">
            <ThemeSwitcher />
        </footer>
    </div>
</template>

<script setup lang="ts">
import JobTable from "@/components/JobTable.vue";
import ThemeSwitcher from "@/components/ThemeSwitcher.vue";
import { SERVER_URL } from "@/main";
import type { JobSource, JobStatus } from "@/types/job";
import { computed, onMounted, onUnmounted, ref } from "vue";

const jobs = ref<JobStatus[]>([]);
const loading = ref(true);
const failed = ref(false);

// Ticks relative times ("5m ago") and running-job runtimes
const now = ref(new Date());
const ticker = setInterval(() => (now.value = new Date()), 1000);

const bySource = (source: JobSource) => jobs.value.filter((j) => j.source === source).sort((a, b) => a.index - b.index);
const githubJobs = computed(() => bySource("github"));
const glueJobs = computed(() => bySource("glue"));

// Server Sent Events Implementation - existing rows stay visible and are replaced in place as fresh statuses arrive
let eventSource: EventSource | null = null;

const loadJobs = (force = false) => {
    if (eventSource) eventSource.close();
    loading.value = true;
    failed.value = false;
    const seen = new Set<string>();

    eventSource = new EventSource(`${SERVER_URL}/jobs-sse${force ? "?refresh=true" : ""}`);
    eventSource.onmessage = (event) => {
        const status: JobStatus = JSON.parse(event.data);
        seen.add(status.id);
        const existing = jobs.value.findIndex((j) => j.id === status.id);
        if (existing >= 0) jobs.value[existing] = status;
        else jobs.value.push(status);
    };
    eventSource.addEventListener("done", () => {
        if (eventSource) eventSource.close();
        jobs.value = jobs.value.filter((j) => seen.has(j.id)); // drop jobs removed from config
        loading.value = false;
    });
    eventSource.onerror = (e) => {
        console.error("EventSource error:", e);
        if (eventSource) eventSource.close();
        failed.value = true;
        loading.value = false;
    };
};

onMounted(() => loadJobs());

onUnmounted(() => {
    if (eventSource) eventSource.close();
    clearInterval(ticker);
});
</script>
