<template>
    <div id="social" class="fixed right-4 bottom-4 z-50">
        <a href="https://github.com/ben-burwood/job-monitor" target="_blank" title="Job Monitor on GitHub">
            <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" class="hover:scale-110 transition-transform">
                <path fill="gray" :d="siGithub.path" />
            </svg>
        </a>
    </div>
</template>

<script setup lang="ts">
import { siGithub } from "simple-icons";
</script>
