<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t("prompts.unarchive") }}</h2>
    </div>

    <div class="card-content">
      <form ref="unarchiveForm">
        <p>{{ $t("prompts.unarchiveFolderNameMessage") }}</p>
        <input
          class="input input--block"
          v-focus
          type="text"
          @keyup.enter="submit"
          v-model.trim="name"
          required
        />
      </form>

      <p>{{ $t("prompts.unarchiveDestinationLocationMessage") }}</p>
      <file-list @update:selected="(val) => (dest = val)"></file-list>

      <p v-if="overwriteAvailable">
        <input type="checkbox" v-model="overwriteExisting" />
        {{ $t("prompts.unarchiveOverwriteExisting") }}
      </p>
    </div>

    <div class="card-action">
      <button
        class="button button--flat button--grey"
        @click="$store.commit('closeHovers')"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')"
      >
        {{ $t("buttons.cancel") }}
      </button>
      <button
        @click="submit"
        class="button button--flat"
        type="submit"
        :aria-label="$t('buttons.unarchive')"
        :title="$t('buttons.unarchive')"
      >
        {{ $t("buttons.unarchive") }}
      </button>
    </div>
  </div>
</template>

<script>
import { mapState, mapGetters } from "vuex";
import FileList from "./FileList.vue";
import { files as api } from "@/api";
import buttons from "@/utils/buttons";

export default {
  name: "unarchive",
  components: { FileList },
  data: function () {
    return {
      overwriteExisting: false,
      dest: null,
      name: "",
    };
  },
  computed: {
    ...mapState(["req", "selected", "selectedCount"]),
    ...mapGetters(["isListing", "isFiles"]),
    overwriteAvailable() {
      let item = this.req.items[this.selected[0]];
      return [
        ".zip",
        ".rar",
        ".tar",
        ".bz2",
        ".gz",
        ".xz",
        ".lz4",
        ".sz",
      ].includes(item.extension);
    },
  },
  methods: {
    cancel: function () {
      this.$store.commit("closeHovers");
    },
    submit: async function () {
      if (!this.$refs.unarchiveForm.reportValidity()) {
        return;
      }

      let item = this.req.items[this.selected[0]];
      let dst = this.dest + encodeURIComponent(this.name);

      try {
        buttons.loading("unarchive");
        this.$store.commit("closeHovers");
        await api.unarchive(item.url, dst, this.overwriteExisting);
        this.$store.commit("setReload", true);
      } catch (e) {
        this.$showError(e);
      } finally {
        buttons.done("unarchive");
      }
      this.$store.dispatch("quota/fetch", 3000);
    },
  },
};
</script>
