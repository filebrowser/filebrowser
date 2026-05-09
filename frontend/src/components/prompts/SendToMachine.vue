<template>
  <div class="card floating">
    <div class="card-content">
      <p>{{ t("prompts.sendToMachineConfirm") }}</p>
      <p class="small">
        <strong>{{ filePath }}</strong>
        <span v-if="lineCount">
          — {{ t("prompts.sendToMachineLines", { n: lineCount }) }}
        </span>
      </p>
      <p class="small">
        {{ t("prompts.sendToMachineDestination") }}
        <strong>{{ haasHostLabel || t("prompts.sendToMachineUnknownHost") }}</strong>
      </p>
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
        class="button button--flat button--blue"
        @click="confirm"
        :aria-label="t('buttons.send')"
        :title="t('buttons.send')"
        tabindex="1"
      >
        {{ t("buttons.send") }}
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
const lineCount = computed(() => (props.value as any).lineCount ?? 0);
const haasHostLabel = computed(() => (props.value as any).haasHostLabel ?? "");

const closeHovers = () => layoutStore.closeHovers();
const confirm = () => layoutStore.currentPrompt?.confirm?.(new Event("submit"));
</script>
