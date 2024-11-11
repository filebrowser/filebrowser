<template>
  <div class="card floating">
    <div class="card-content">
      <p v-if="!this.isListing || selectedCount === 1">
        {{ $t("prompts.deleteMessageSingle") }}
      </p>
      <p v-else>
        {{ $t("prompts.deleteMessageMultiple", { count: selectedCount }) }}
      </p>
      <p v-if="trashBinCheckbox">
        <input type="checkbox" v-model="skipTrash" />
        {{ $t("prompts.skipTrashMessage") }}
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

<script>
import { mapActions, mapState, mapWritableState } from "pinia";
import { files as api } from "@/api";
import buttons from "@/utils/buttons";
import { trashDir } from "@/utils/constants";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { useQuotaStore } from "@/stores/quota";

export default {
  name: "delete",
  data: function () {
    return {
      skipTrash: true,
    };
  },
  inject: ["$showError"],
  computed: {
    ...mapState(useFileStore, [
      "isListing",
      "selectedCount",
      "req",
      "selected",
      "currentPrompt",
    ]),
    ...mapWritableState(useFileStore, ["reload"]),
    trashBinCheckbox() {
      if (trashDir === "") {
        return false;
      }

      if (
        this.req.path.startsWith(`/${trashDir}/`) ||
        this.req.path.endsWith(`/${trashDir}`)
      ) {
        return false;
      }

      if (
        this.selectedCount == 1 &&
        this.req.items[this.selected].name == trashDir
      ) {
        return false;
      }

      return true;
    },
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers"]),
    ...mapActions(useQuotaStore, ["fetchQuota"]),
    submit: async function () {
      buttons.loading("delete");

      window.sessionStorage.setItem("modified", "true");
      try {
        if (!this.isListing) {
          await api.remove(this.$route.path, this.skipTrash);
          buttons.success("delete");

          this.currentPrompt?.confirm();
          this.closeHovers();
          return;
        }

        this.closeHovers();

        if (this.selectedCount === 0) {
          return;
        }

        let promises = [];
        for (let index of this.selected) {
          promises.push(api.remove(this.req.items[index].url, this.skipTrash));
        }

        await Promise.all(promises);
        buttons.success("delete");
        this.reload = true;
        this.fetchQuota(3000);
      } catch (e) {
        buttons.done("delete");
        this.$showError(e);
        if (this.isListing) this.reload = true;
      }
    },
  },
};
</script>
