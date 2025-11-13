<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t("prompts.replace") }}</h2>
    </div>

    <div class="card-content">
      <p>{{ $t("prompts.replaceMessage") }}</p>
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
        @click="(event) => currentPrompt?.confirm(event, 'rename')"
        :aria-label="$t('buttons.rename')"
        :title="$t('buttons.rename')"
        tabindex="2"
      >
        {{ $t("buttons.rename") }}
      </button>
      <button
        id="focus-prompt"
        class="button button--flat button--red"
        @click="(event) => currentPrompt?.confirm(event, 'overwrite')"
        :aria-label="$t('buttons.replace')"
        :title="$t('buttons.replace')"
        tabindex="1"
      >
        {{ $t("buttons.replace") }}
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
