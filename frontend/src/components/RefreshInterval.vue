<template>
    <div ref="root" class="relative">
        <button
            type="button"
            class="flex items-center gap-1.5 px-3 py-1.5 rounded-full hover:bg-base-200 transition-colors"
            :aria-label="`Refresh interval: ${formatRefreshInterval(refreshInterval)}`"
            :aria-expanded="showRefreshMenu"
            @click="showRefreshMenu = !showRefreshMenu"
        >
            <svg
                class="h-3.5 w-3.5 text-base-content/60"
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
            >
                <path d="M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8" />
                <path d="M21 3v5h-5" />
                <path d="M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16" />
                <path d="M8 16H3v5" />
            </svg>
            <span class="text-xs font-medium">{{ formatRefreshInterval(refreshInterval) }}</span>
        </button>

        <!-- refresh interval dropdown -->
        <div v-if="showRefreshMenu" class="absolute bottom-full left-0 mb-2 bg-base-100 border border-base-300 rounded-lg shadow-lg overflow-hidden">
            <button
                v-for="interval in REFRESH_INTERVALS"
                :key="interval.value"
                type="button"
                class="block w-full px-4 py-2 text-xs text-left hover:bg-base-200 transition-colors"
                :class="{ 'bg-base-200': refreshInterval === interval.value }"
                @click="selectRefreshInterval(interval.value)"
            >
                {{ interval.label }}
            </button>
        </div>
    </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from "vue";

const emit = defineEmits<{ refresh: [] }>();

const HOUR = 60 * 60;
const REFRESH_INTERVALS = [
    { value: HOUR, label: "1h" },
    { value: 6 * HOUR, label: "6h" },
    { value: 12 * HOUR, label: "12h" },
    { value: 24 * HOUR, label: "24h" },
];
const DEFAULT_REFRESH_INTERVAL = HOUR;
const REFRESH_INTERVAL_STORAGE_KEY = "job-monitor:refresh-interval";

const getStoredRefreshInterval = () => {
    const stored = Number(localStorage.getItem(REFRESH_INTERVAL_STORAGE_KEY));
    return REFRESH_INTERVALS.some((i) => i.value === stored) ? stored : DEFAULT_REFRESH_INTERVAL;
};

const root = ref<HTMLElement | null>(null);
const refreshInterval = ref(getStoredRefreshInterval());
const showRefreshMenu = ref(false);
let refreshIntervalHandler: ReturnType<typeof setInterval> | null = null;

const formatRefreshInterval = (seconds: number) => REFRESH_INTERVALS.find((i) => i.value === seconds)?.label ?? `${seconds}s`;

const setRefreshInterval = (seconds: number) => {
    localStorage.setItem(REFRESH_INTERVAL_STORAGE_KEY, String(seconds));
    if (refreshIntervalHandler) clearInterval(refreshIntervalHandler);
    refreshIntervalHandler = setInterval(() => emit("refresh"), seconds * 1000);
};

const selectRefreshInterval = (seconds: number) => {
    refreshInterval.value = seconds;
    showRefreshMenu.value = false;
    setRefreshInterval(seconds);
};

// Close the refresh menu when clicking outside
const handleClickOutside = (event: MouseEvent) => {
    if (root.value && !root.value.contains(event.target as Node)) showRefreshMenu.value = false;
};

onMounted(() => {
    setRefreshInterval(refreshInterval.value);
    document.addEventListener("click", handleClickOutside);
});

onUnmounted(() => {
    if (refreshIntervalHandler) clearInterval(refreshIntervalHandler);
    document.removeEventListener("click", handleClickOutside);
});
</script>
