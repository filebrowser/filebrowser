<template>
  <div class="quota-usage" v-if="quotaLimit > 0">
    <h3>{{ t("settings.quotaUsage") }}</h3>
    <div class="quota-bar-container">
      <div class="quota-bar">
        <div
          class="quota-fill"
          :class="usageClass"
          :style="{ width: usagePercentage + '%' }"
        ></div>
      </div>
      <div class="quota-text">
        <span>{{ formatSize(quotaUsed) }} / {{ formatSize(quotaLimit) }}</span>
        <span class="quota-percentage" :class="usageClass">
          {{ usagePercentage.toFixed(1) }}%
        </span>
      </div>
    </div>
    <div v-if="usagePercentage >= 95" class="quota-warning critical">
      <i class="material-icons">warning</i>
      {{ t("settings.quotaExceeded") }}
    </div>
    <div v-else-if="usagePercentage >= 80" class="quota-warning">
      <i class="material-icons">info</i>
      {{ t("settings.quotaWarning") }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();

interface Props {
  quotaUsed: number;
  quotaLimit: number;
  quotaUnit?: string;
}

const props = withDefaults(defineProps<Props>(), {
  quotaUnit: "GB",
});

const usagePercentage = computed(() => {
  if (props.quotaLimit === 0) return 0;
  return Math.min((props.quotaUsed / props.quotaLimit) * 100, 100);
});

const usageClass = computed(() => {
  const percentage = usagePercentage.value;
  if (percentage >= 95) return "critical";
  if (percentage >= 80) return "warning";
  return "normal";
});

function formatSize(bytes: number): string {
  const TB = 1024 * 1024 * 1024 * 1024;
  const GB = 1024 * 1024 * 1024;
  const MB = 1024 * 1024;
  const KB = 1024;

  if (bytes >= TB) {
    return `${(bytes / TB).toFixed(2)} TB`;
  }
  if (bytes >= GB) {
    return `${(bytes / GB).toFixed(2)} GB`;
  }
  if (bytes >= MB) {
    return `${(bytes / MB).toFixed(2)} MB`;
  }
  if (bytes >= KB) {
    return `${(bytes / KB).toFixed(2)} KB`;
  }
  return `${bytes} B`;
}
</script>

<style scoped>
.quota-usage {
  margin: 1rem 0;
  padding: 1rem;
  background-color: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 4px;
}

.quota-usage h3 {
  margin: 0 0 1rem 0;
  font-size: 1rem;
  font-weight: 500;
}

.quota-bar-container {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.quota-bar {
  width: 100%;
  height: 24px;
  background-color: var(--background-hover);
  border-radius: 12px;
  overflow: hidden;
  position: relative;
}

.quota-fill {
  height: 100%;
  transition: width 0.3s ease, background-color 0.3s ease;
  border-radius: 12px;
}

.quota-fill.normal {
  background-color: #4caf50;
}

.quota-fill.warning {
  background-color: #ff9800;
}

.quota-fill.critical {
  background-color: #f44336;
}

.quota-text {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.9rem;
  color: var(--text-color);
}

.quota-percentage {
  font-weight: 600;
}

.quota-percentage.normal {
  color: #4caf50;
}

.quota-percentage.warning {
  color: #ff9800;
}

.quota-percentage.critical {
  color: #f44336;
}

.quota-warning {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 0.75rem;
  padding: 0.5rem;
  background-color: #fff3cd;
  color: #856404;
  border: 1px solid #ffeaa7;
  border-radius: 4px;
  font-size: 0.85rem;
}

.quota-warning.critical {
  background-color: #f8d7da;
  color: #721c24;
  border-color: #f5c6cb;
}

.quota-warning i {
  font-size: 1.2rem;
}
</style>
