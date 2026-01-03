<template>
  <div class="quota-input">
    <h3>{{ t("settings.quota") }}</h3>
    <div class="quota-controls">
      <div class="input-group">
        <label>
          <input
            type="checkbox"
            v-model="localEnforceQuota"
            @change="updateQuota"
          />
          {{ t("settings.enforceQuota") }}
        </label>
      </div>
      <div class="input-group" v-if="localEnforceQuota">
        <label>{{ t("settings.quotaLimit") }}</label>
        <div class="quota-input-row">
          <input
            type="number"
            v-model.number="localQuotaValue"
            @input="updateQuota"
            min="0"
            step="0.1"
            :disabled="!localEnforceQuota"
          />
          <select
            v-model="localQuotaUnit"
            @change="updateQuota"
            :disabled="!localEnforceQuota"
          >
            <option value="KB">KB</option>
            <option value="MB">MB</option>
            <option value="GB">GB</option>
            <option value="TB">TB</option>
          </select>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();

interface Props {
  quotaLimit?: number;
  quotaUnit?: string;
  enforceQuota?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  quotaLimit: 0,
  quotaUnit: "GB",
  enforceQuota: false,
});

const emit = defineEmits<{
  (e: "update:quotaLimit", value: number): void;
  (e: "update:quotaUnit", value: string): void;
  (e: "update:enforceQuota", value: boolean): void;
}>();

const localQuotaValue = ref(0);
const localQuotaUnit = ref("GB");
const localEnforceQuota = ref(false);

const KB = 1024;
const MB = 1024 * KB;
const GB = 1024 * MB;
const TB = 1024 * GB;

function convertFromBytes(bytes: number, unit: string): number {
  switch (unit) {
    case "KB":
      return bytes / KB;
    case "MB":
      return bytes / MB;
    case "TB":
      return bytes / TB;
    case "GB":
    default:
      return bytes / GB;
  }
}

function convertToBytes(value: number, unit: string): number {
  switch (unit) {
    case "KB":
      return value * KB;
    case "MB":
      return value * MB;
    case "TB":
      return value * TB;
    case "GB":
    default:
      return value * GB;
  }
}

function updateQuota() {
  emit("update:enforceQuota", localEnforceQuota.value);
  emit("update:quotaUnit", localQuotaUnit.value);
  
  if (localEnforceQuota.value) {
    const bytes = convertToBytes(localQuotaValue.value, localQuotaUnit.value);
    emit("update:quotaLimit", bytes);
  } else {
    emit("update:quotaLimit", 0);
  }
}

watch(
  () => [props.quotaLimit, props.quotaUnit, props.enforceQuota],
  () => {
    localEnforceQuota.value = props.enforceQuota;
    localQuotaUnit.value = props.quotaUnit || "GB";
    
    if (props.quotaLimit > 0) {
      localQuotaValue.value = convertFromBytes(
        props.quotaLimit,
        localQuotaUnit.value
      );
    } else {
      localQuotaValue.value = 0;
    }
  },
  { immediate: true }
);

onMounted(() => {
  localEnforceQuota.value = props.enforceQuota;
  localQuotaUnit.value = props.quotaUnit || "GB";
  
  if (props.quotaLimit > 0) {
    localQuotaValue.value = convertFromBytes(
      props.quotaLimit,
      localQuotaUnit.value
    );
  }
});
</script>

<style scoped>
.quota-input {
  margin: 1rem 0;
}

.quota-input h3 {
  margin-bottom: 0.5rem;
  font-size: 1rem;
  font-weight: 500;
}

.quota-controls {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.input-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.input-group label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.9rem;
}

.quota-input-row {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.quota-input-row input[type="number"] {
  flex: 1;
  padding: 0.5rem;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-size: 0.9rem;
}

.quota-input-row select {
  padding: 0.5rem;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-size: 0.9rem;
  min-width: 80px;
}

input[type="checkbox"] {
  cursor: pointer;
}

input:disabled,
select:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
