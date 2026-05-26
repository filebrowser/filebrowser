<template>
  <div class="card floating">
    <div class="card-content">
      <p v-if="!isListing || selectedCount === 1">
        {{ $t("prompts.deleteMessageSingle", { name: deleteName }) }}
      </p>
      <template v-else>
        <p>
          {{ $t("prompts.deleteMessageMultiple", { count: selectedCount }) }}
        </p>
        <details>
          <summary>{{ $t("prompts.showFiles") }}</summary>
          <ul class="delete-file-list">
            <li v-for="index in selected" :key="index">
              {{ req.items[index].name }}
            </li>
          </ul>
        </details>
      </template>
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
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";

export default {
  name: "delete",
  inject: ["$showError"],
  computed: {
    ...mapState(useFileStore, [
      "isListing",
      "selectedCount",
      "req",
      "selected",
    ]),
    ...mapState(useLayoutStore, ["currentPrompt"]),
    ...mapWritableState(useFileStore, ["reload", "preselect"]),
    deleteName() {
      if (this.isListing && this.selectedCount === 1) {
        return this.req.items[this.selected[0]].name;
      }
      return this.req?.name || "";
    },
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers"]),
    submit: async function () {
      buttons.loading("delete");

      try {
        if (!this.isListing) {
          await api.remove(this.$route.path);
          buttons.success("delete");

          this.currentPrompt?.confirm();
          this.closeHovers();
          return;
        }

        this.closeHovers();

        if (this.selectedCount === 0) {
          return;
        }

        const promises = [];
        for (const index of this.selected) {
          promises.push(api.remove(this.req.items[index].url));
        }

        await Promise.all(promises);
        buttons.success("delete");

        const nearbyItem =
          this.req.items[Math.max(0, Math.min(this.selected) - 1)];

        this.preselect = nearbyItem?.path;

        this.reload = true;
      } catch (e) {
        buttons.done("delete");
        this.$showError(e);
        if (this.isListing) this.reload = true;
      }
    },
  },
};
</script>

<style scoped>
.delete-file-list {
  max-height: 10em;
  overflow-y: auto;
  margin: 0.5em 0 0;
  padding-left: 1.5em;
}

details summary {
  cursor: pointer;
  margin-top: 0.5em;
  color: var(--blue);
}
</style>
