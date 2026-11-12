<template>
    <div id="settings" class="fixed bottom-4 left-4 z-50">
        <div class="flex items-center gap-1 bg-base-100/95 backdrop-blur-sm border border-base-300 rounded-full shadow-md p-1">
            <RefreshInterval @refresh="$emit('refresh')" />

            <!-- divider -->
            <div class="h-5 w-px bg-base-300" />

            <!-- theme toggle -->
            <div class="relative group">
                <button
                    type="button"
                    class="swap swap-rotate p-1.5 rounded-full hover:bg-base-200 transition-colors"
                    :aria-label="darkMode ? 'Switch to light mode' : 'Switch to dark mode'"
                    data-toggle-theme="dark,light"
                    data-act-class="swap-active"
                >
                    <!-- sun icon (shown in dark mode) -->
                    <svg
                        class="swap-on h-3.5 w-3.5"
                        xmlns="http://www.w3.org/2000/svg"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                    >
                        <circle cx="12" cy="12" r="4" />
                        <path d="M12 2v2" />
                        <path d="M12 20v2" />
                        <path d="m4.93 4.93 1.41 1.41" />
                        <path d="m17.66 17.66 1.41 1.41" />
                        <path d="M2 12h2" />
                        <path d="M20 12h2" />
                        <path d="m6.34 17.66-1.41 1.41" />
                        <path d="m19.07 4.93-1.41 1.41" />
                    </svg>

                    <!-- moon icon (shown in light mode) -->
                    <svg
                        class="swap-off h-3.5 w-3.5"
                        xmlns="http://www.w3.org/2000/svg"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                    >
                        <path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z" />
                    </svg>
                </button>

                <!-- tooltip -->
                <div
                    class="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 px-2 py-1 bg-base-100 border border-base-300 text-xs rounded-md shadow-md opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none whitespace-nowrap"
                >
                    {{ darkMode ? "Light mode" : "Dark mode" }}
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { themeChange } from "theme-change";
import { onMounted, onUnmounted, ref } from "vue";
import RefreshInterval from "./RefreshInterval.vue";

defineEmits<{ refresh: [] }>();

const isDark = () => document.documentElement.getAttribute("data-theme") === "dark";

const darkMode = ref(isDark());

// theme-change updates data-theme directly, so watch it to keep the tooltip in sync
const observer = new MutationObserver(() => {
    darkMode.value = isDark();
});

onMounted(() => {
    themeChange(false);
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ["data-theme"] });
});

onUnmounted(() => {
    observer.disconnect();
});
</script>

<style scoped>
@keyframes slideIn {
    from {
        transform: translateX(-20px);
        opacity: 0;
    }
    to {
        transform: translateX(0);
        opacity: 1;
    }
}

#settings {
    animation: slideIn 0.3s ease-out;
}

#settings > div {
    transition: all 0.2s ease;
}

#settings > div:hover {
    transform: translateY(-2px);
    box-shadow:
        0 10px 25px -5px rgb(0 0 0 / 0.1),
        0 8px 10px -6px rgb(0 0 0 / 0.1);
}
</style>
