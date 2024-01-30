<template>
  <div class="card floating">
    <div class="card-content">
      <p>
        {{ $t("prompts.discardEditorChanges") }}
      </p>
    </div>
    <div class="card-action">
      <button
        @click="$store.commit('closeHovers')"
        class="button button--flat button--grey"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')"
      >
        {{ $t("buttons.cancel") }}
      </button>
      <button
        @click="submit"
        class="button button--flat button--red"
        :aria-label="$t('buttons.discardChanges')"
        :title="$t('buttons.discardChanges')"
      >
        {{ $t("buttons.discardChanges") }}
      </button>
    </div>
  </div>
</template>

<script>
import { mapMutations } from "vuex";
import url from "@/utils/url";

export default {
  name: "discardEditorChanges",
  methods: {
    ...mapMutations(["closeHovers"]),
    submit: async function () {
      this.$store.commit("updateRequest", {});

      let uri = url.removeLastDir(this.$route.path) + "/";
      this.$router.push({ path: uri });
    },
  },
};
</script>
