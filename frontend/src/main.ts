import { createApp } from 'vue';
import App from './App.vue';
import './style.css';

export const SERVER_URL = `${window.location.protocol}//${window.location.hostname}:${window.location.port}/api`;

createApp(App).mount('#app')
