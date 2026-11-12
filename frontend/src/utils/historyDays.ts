import { ref, watch } from "vue";

// Selectable run history lengths; must match jobs.HistoryDays on the server
export const HISTORY_DAYS = [1, 7, 14, 28];
const DEFAULT_HISTORY_DAYS = 7;
const HISTORY_DAYS_STORAGE_KEY = "job-monitor:history-days";

const getStoredHistoryDays = () => {
  const stored = Number(localStorage.getItem(HISTORY_DAYS_STORAGE_KEY));
  return HISTORY_DAYS.includes(stored) ? stored : DEFAULT_HISTORY_DAYS;
};

// How many days back run histories reach, shared by the setting and every open history
export const historyDays = ref(getStoredHistoryDays());

watch(historyDays, (days) => localStorage.setItem(HISTORY_DAYS_STORAGE_KEY, String(days)));
