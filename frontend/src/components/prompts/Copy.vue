<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t("prompts.copy") }}</h2>
    </div>

    <div class="card-content">
      <p>{{ $t("prompts.copyMessage") }}</p>
      <file-list ref="fileList" @update:selected="(val) => (dest = val)">
      </file-list>
    </div>

    <div
      class="card-action"
      style="display: flex; align-items: center; justify-content: space-between;"
    >
      <template v-if="user.perm.create">
        <button
          class="button button--flat"
          @click="$refs.fileList.createDir()"
          :aria-label="$t('sidebar.newFolder')"
          :title="$t('sidebar.newFolder')"
          style="justify-self: left;"
        >
          <span>{{ $t("sidebar.newFolder") }}</span>
        </button>
      </template>
      <div>
        <button
          class="button button--flat button--grey"
          @click="$store.commit('closeHovers')"
          :aria-label="$t('buttons.cancel')"
          :title="$t('buttons.cancel')"
        >
          {{ $t("buttons.cancel") }}
        </button>
        <button
          class="button button--flat"
          @click="copy"
          :aria-label="$t('buttons.copy')"
          :title="$t('buttons.copy')"
        >
          {{ $t("buttons.copy") }}
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import { mapState } from "vuex";
import FileList from "./FileList.vue";
import { files as api } from "@/api";
import buttons from "@/utils/buttons";
import * as upload from "@/utils/upload";

export default {
  name: "copy",
  components: { FileList },
  data: function () {
    return {
      current: window.location.pathname,
      dest: null,
    };
  },
  computed: mapState(["req", "selected", "user"]),
  methods: {
    copy: async function (event) {
      event.preventDefault();
      let items = [];

      // Create a new promise for each file.
      for (let item of this.selected) {
        items.push({
          from: this.req.items[item].url,
          to: this.dest + encodeURIComponent(this.req.items[item].name),
          name: this.req.items[item].name,
        });
      }

      let action = async (overwrite, rename) => {
        buttons.loading("copy");

        await api
          .copy(items, overwrite, rename)
          .then(() => {
            buttons.success("copy");

            if (this.$route.path === this.dest) {
              this.$store.commit("setReload", true);

              return;
            }

            this.$router.push({ path: this.dest });
          })
          .catch((e) => {
            buttons.done("copy");
            this.$showError(e);
          });
      };

      if (this.$route.path === this.dest) {
        this.$store.commit("closeHovers");
        action(false, true);
        this.$store.dispatch("quota/fetch", 3000);
        return;
      }

      let dstItems = (await api.fetch(this.dest)).items;
      let conflict = upload.checkConflict(items, dstItems);

      let overwrite = false;
      let rename = false;

      if (conflict) {
        this.$store.commit("showHover", {
          prompt: "replace-rename",
          confirm: (event, option) => {
            overwrite = option == "overwrite";
            rename = option == "rename";

            event.preventDefault();
            this.$store.commit("closeHovers");
            action(overwrite, rename);
            this.$store.dispatch("quota/fetch", 3000);
          },
        });

        return;
      }

      action(overwrite, rename);
      this.$store.dispatch("quota/fetch", 3000);
    },
  },
};
</script>
