<template>
  <div class="card floating">
    <div class="card-content">
      <p v-if="req.kind !== 'listing'">
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
        :aria-label="$t('buttons.delete')"
        :title="$t('buttons.delete')"
      >
        {{ $t("buttons.delete") }}
      </button>
    </div>
  </div>
</template>

<script>
import { mapGetters, mapMutations, mapState } from "vuex";
import { files as api } from "@/api";
import buttons from "@/utils/buttons";
import { trashDir } from "@/utils/constants";

export default {
  name: "delete",
  data: function () {
    return {
      skipTrash: true,
    };
  },
  computed: {
    ...mapGetters(["isListing", "selectedCount"]),
    ...mapState(["req", "selected", "showConfirm"]),
    trashBinCheckbox() {
      if (trashDir === "") {
        return false;
      }

      if (this.req.path.startsWith(`/${trashDir}/`)) {
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
    ...mapMutations(["closeHovers"]),
    submit: async function () {
      buttons.loading("delete");

      try {
        if (!this.isListing) {
          await api.remove(this.$route.path, this.skipTrash);
          buttons.success("delete");

          this.showConfirm();
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
        this.$store.commit("setReload", true);
      } catch (e) {
        buttons.done("delete");
        this.$showError(e);
        if (this.isListing) this.$store.commit("setReload", true);
      }
    },
  },
};
</script>
