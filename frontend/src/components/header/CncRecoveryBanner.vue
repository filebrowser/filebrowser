<template>
  <div v-if="cnc.recoveryPending" class="recovery-banner" role="alert">
    <i class="material-icons">warning</i>
    <span class="recovery-banner__text">
      {{
        t("cnc.recoveryBanner", {
          path: cnc.recoveryFilePath || t("cnc.recoveryUnknownFile"),
        })
      }}
    </span>
    <button
      class="recovery-banner__btn"
      :disabled="acking"
      @click="ack"
      :title="t('cnc.recoveryAck')"
    >
      {{ acking ? t("cnc.recoveryAcking") : t("cnc.recoveryAck") }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, ref } from "vue";
import { useI18n } from "vue-i18n";
import { useCncStore } from "@/stores/cnc";

const { t } = useI18n();
const store = useCncStore();
const cnc = computed(() => store);

const $showError = inject<IToastError>("$showError")!;

const acking = ref(false);
const ack = async () => {
  if (acking.value) return;
  acking.value = true;
  try {
    await store.ackRecovery();
  } catch (e: any) {
    $showError(e);
  } finally {
    acking.value = false;
  }
};
</script>

<style>
.recovery-banner {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.5rem 0.9rem;
  background: #c97a00;
  color: #fff;
  font-size: 0.9rem;
}

.recovery-banner__text {
  flex: 1;
  min-width: 0;
}

.recovery-banner__btn {
  background: rgba(255, 255, 255, 0.18);
  color: #fff;
  border: 1px solid rgba(255, 255, 255, 0.4);
  border-radius: 4px;
  padding: 0.25rem 0.7rem;
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
}

.recovery-banner__btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.3);
}

.recovery-banner__btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
