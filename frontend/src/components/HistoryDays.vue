<template>
    <div ref="root" class="relative">
        <button
            type="button"
            class="flex items-center gap-1.5 px-3 py-1.5 rounded-full hover:bg-base-200 transition-colors"
            :aria-label="`Run history: ${historyDays} days`"
            :aria-expanded="showMenu"
            title="Run history length"
            @click="showMenu = !showMenu"
        >
            <IconCalendar class="h-3.5 w-3.5 text-base-content/60" />
            <span class="text-xs font-medium">{{ historyDays }}d</span>
        </button>

        <!-- history days dropdown -->
        <div v-if="showMenu" class="absolute bottom-full left-0 mb-2 bg-base-100 border border-base-300 rounded-lg shadow-lg overflow-hidden">
            <button
                v-for="days in HISTORY_DAYS"
                :key="days"
                type="button"
                class="block w-full px-4 py-2 text-xs text-left hover:bg-base-200 transition-colors"
                :class="{ 'bg-base-200': historyDays === days }"
                @click="selectHistoryDays(days)"
            >
                {{ days }}d
            </button>
        </div>
    </div>
</template>

<script setup lang="ts">
import { HISTORY_DAYS, historyDays } from "@/utils/historyDays";
import { onMounted, onUnmounted, ref } from "vue";
import IconCalendar from "~icons/lucide/calendar";

const root = ref<HTMLElement | null>(null);
const showMenu = ref(false);

const selectHistoryDays = (days: number) => {
    historyDays.value = days;
    showMenu.value = false;
};

// Close the menu when clicking outside
const handleClickOutside = (event: MouseEvent) => {
    if (root.value && !root.value.contains(event.target as Node)) showMenu.value = false;
};

onMounted(() => document.addEventListener("click", handleClickOutside));
onUnmounted(() => document.removeEventListener("click", handleClickOutside));
</script>
