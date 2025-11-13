<template>
  <div class="card floating">
    <div class="card-content">
      <p>{{ $t("prompts.deleteMessageShare", { path: "" }) }}</p>
    </div>
    <div class="card-action">
      <button
        @click="closeHovers"
        class="button button--flat button--grey"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')"
        tabindex="2"
      >
        {{ $t("buttons.cancel") }}
      </button>
      <button
        id="focus-prompt"
        @click="submit"
        class="button button--flat button--red"
        :aria-label="$t('buttons.delete')"
        :title="$t('buttons.delete')"
        tabindex="1"
      >
        {{ $t("buttons.delete") }}
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

const submit = () => {
  currentPrompt.value?.confirm();
};
</script>
