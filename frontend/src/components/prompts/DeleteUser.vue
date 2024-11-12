<template>
  <div class="card floating">
    <div class="card-content">
      <p>{{ t("prompts.deleteUser") }}</p>
    </div>

    <div class="card-action">
      <button
        id="focus-prompt"
        class="button button--flat button--grey"
        @click="layoutStore.closeHovers"
        :aria-label="t('buttons.cancel')"
        :title="t('buttons.cancel')"
        tabindex="1"
      >
        {{ t("buttons.cancel") }}
      </button>
      <button
        class="button button--flat"
        @click="layoutStore.currentPrompt?.confirm"
        tabindex="2"
      >
        {{ t("buttons.delete") }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useLayoutStore } from "@/stores/layout";
import { useI18n } from "vue-i18n";

const layoutStore = useLayoutStore();

const { t } = useI18n();

// const emit = defineEmits<{
//   (e: "confirm"): void;
// }>();
</script>
