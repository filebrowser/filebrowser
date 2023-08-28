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
      >
        {{ $t("buttons.cancel") }}
      </button>
      <button
        class="button button--flat button--blue"
        @click="showAction"
        :aria-label="$t('buttons.continue')"
        :title="$t('buttons.continue')"
      >
        {{ $t("buttons.continue") }}
      </button>
      <button
        class="button button--flat button--red"
        @click="showConfirm"
        :aria-label="$t('buttons.replace')"
        :title="$t('buttons.replace')"
      >
        {{ $t("buttons.replace") }}
      </button>
    </div>
  </div>
</template>

<script>
import { mapActions, mapState } from "pinia";
import { useLayoutStore } from "@/stores/layout";

export default {
  name: "replace",
  computed: {
    ...mapState(useLayoutStore, ["showConfirm", "showAction"]),
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers"]),
  },
};
</script>
