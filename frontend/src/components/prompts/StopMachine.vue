<template>
  <div class="card floating">
    <div class="card-content">
      <p>{{ t("prompts.stopMachineConfirm") }}</p>
      <p class="small" v-if="filePath">
        <strong>{{ filePath }}</strong>
        <span v-if="lineCurrent > 0">
          — {{ t("prompts.stopMachineAtLine", { n: lineCurrent }) }}
        </span>
      </p>
      <p class="small">{{ t("prompts.stopMachineHint") }}</p>
    </div>
    <div class="card-action">
      <button
        class="button button--flat button--grey"
        @click="closeHovers"
        :aria-label="t('buttons.cancel')"
        :title="t('buttons.cancel')"
        tabindex="2"
      >
        {{ t("buttons.cancel") }}
      </button>
      <button
        id="focus-prompt"
        class="button button--flat button--red"
        @click="confirm"
        :aria-label="t('buttons.stopMachine')"
        :title="t('buttons.stopMachine')"
        tabindex="1"
      >
        {{ t("buttons.stopMachine") }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useLayoutStore } from "@/stores/layout";

const { t } = useI18n();
const layoutStore = useLayoutStore();

const props = computed(() => layoutStore.currentPrompt?.props || {});
const filePath = computed(() => (props.value as any).filePath ?? "");
const lineCurrent = computed(() => (props.value as any).lineCurrent ?? 0);

const closeHovers = () => layoutStore.closeHovers();
const confirm = () => layoutStore.currentPrompt?.confirm?.(new Event("submit"));
</script>
