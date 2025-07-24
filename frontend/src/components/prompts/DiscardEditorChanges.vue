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
        tabindex="2"
      >
        {{ $t("buttons.cancel") }}
      </button>
      <button
        id="focus-prompt"
        @click="currentPrompt.confirm"
        class="button button--flat button--red"
        :aria-label="$t('buttons.discardChanges')"
        :title="$t('buttons.discardChanges')"
        tabindex="1"
      >
        {{ $t("buttons.discardChanges") }}
      </button>
    </div>
  </div>
</template>

<script>
import { mapState, mapActions } from "pinia";
import { useLayoutStore } from "@/stores/layout";

export default {
  name: "discardEditorChanges",
  computed: {
    ...mapState(useLayoutStore, ["currentPrompt"]),
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers"]),
  },
};
</script>
