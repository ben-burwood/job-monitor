<template>
    <span class="badge whitespace-nowrap" :class="STATE_CLASSES[job.state].badge" :title="job.error">
        <span v-if="job.state === 'running'" class="loading loading-spinner loading-xs"></span>
        {{ job.label }}
    </span>
</template>

<script setup lang="ts">
import type { JobRun } from "@/types/job";
import { STATE_CLASSES } from "@/utils/state";

defineProps<{
    job: JobRun & { error?: string };
}>();
</script>
