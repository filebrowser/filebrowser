<template>
  <div class="card floating">
    <div class="card-content">
      <p>
        {{ $t("prompts.publishMessage") }}
      </p>
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
        class="button button--flat button--blue"
        :aria-label="$t('buttons.publish')"
        :title="$t('buttons.publish')"
        tabindex="1"
      >
        {{ $t("buttons.publish") }}
      </button>
    </div>
  </div>
</template>

<script>
import { mapActions, mapState } from "pinia";
import { torrent as api } from "@/api";
import buttons from "@/utils/buttons";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";

export default {
  name: "publish",
  inject: ["$showError", "$showSuccess"],
  computed: {
    ...mapState(useFileStore, [
      "isListing",
      "selectedCount",
      "req",
      "selected",
      "currentPrompt",
    ]),
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers"]),
    submit: async function () {

      try {
        this.closeHovers();

        if (this.selectedCount === 0) {
          return;
        }

        let action = async () => {
          buttons.loading("publish");
          const res = await api.publish(
            this.req.items[this.selected[0]].url
          ).then(() => {
            buttons.success("publish");
            this.$showSuccess(this.$t("success.torrentPublished"));
          }).catch((e) => {
            buttons.done("publish");
            this.$showError(e);
          });
        };

        action();
      } catch (e) {
        this.$showError(e);
      }
    },
  },
};
</script>
