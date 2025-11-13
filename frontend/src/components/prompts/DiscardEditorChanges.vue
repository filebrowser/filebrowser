<template>
  <div class="card floating">
    <div class="card-content">
      <p>
        {{ $t("prompts.discardEditorChanges") }}
      </p>
    </div>
    <div class="card-action">
      <button
        class="button button--flat button--grey"
        @click="closeHovers"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')"
        tabindex="3"
      >
        {{ $t("buttons.cancel") }}
      </button>
      <button
        class="button button--flat button--blue"
        @click="currentPrompt?.saveAction"
        :aria-label="$t('buttons.saveChanges')"
        :title="$t('buttons.saveChanges')"
        tabindex="1"
      >
        {{ $t("buttons.saveChanges") }}
      </button>
      <button
        id="focus-prompt"
        @click="currentPrompt?.confirm"
        class="button button--flat button--red"
        :aria-label="$t('buttons.discardChanges')"
        :title="$t('buttons.discardChanges')"
        tabindex="2"
      >
        {{ $t("buttons.discardChanges") }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { storeToRefs } from "pinia";
import { useLayoutStore } from "@/stores/layout";

const layoutStore = useLayoutStore();
const { currentPrompt } = storeToRefs(layoutStore);
const { closeHovers } = layoutStore;
</script>
