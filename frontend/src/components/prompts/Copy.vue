<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t("prompts.copy") }}</h2>
    </div>

    <div class="card-content">
      <p>{{ $t("prompts.copyMessage") }}</p>
      <file-list
        ref="fileList"
        @update:selected="(val) => (dest = val)"
        tabindex="1"
      />
    </div>

    <div
      class="card-action"
      style="display: flex; align-items: center; justify-content: space-between"
    >
      <template v-if="user.perm.create">
        <button
          class="button button--flat"
          @click="$refs.fileList.createDir()"
          :aria-label="$t('sidebar.newFolder')"
          :title="$t('sidebar.newFolder')"
          style="justify-self: left"
        >
          <span>{{ $t("sidebar.newFolder") }}</span>
        </button>
      </template>
      <div>
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
          id="focus-prompt"
          class="button button--flat"
          @click="copy"
          :aria-label="$t('buttons.copy')"
          :title="$t('buttons.copy')"
          tabindex="2"
        >
          {{ $t("buttons.copy") }}
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import { mapActions, mapState, mapWritableState } from "pinia";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { useAuthStore } from "@/stores/auth";
import FileList from "./FileList.vue";
import { files as api } from "@/api";
import buttons from "@/utils/buttons";
import * as upload from "@/utils/upload";
import { removePrefix } from "@/api/utils";

export default {
  name: "copy",
  components: { FileList },
  data: function () {
    return {
      current: window.location.pathname,
      dest: null,
    };
  },
  inject: ["$showError"],
  computed: {
    ...mapState(useFileStore, ["req", "selected"]),
    ...mapState(useAuthStore, ["user"]),
    ...mapWritableState(useFileStore, ["reload", "preselect"]),
  },
  methods: {
    ...mapActions(useLayoutStore, ["showHover", "closeHovers"]),
    copy: async function (event) {
      event.preventDefault();
      const items = [];

      // Create a new promise for each file.
      for (const item of this.selected) {
        items.push({
          from: this.req.items[item].url,
          to: this.dest + encodeURIComponent(this.req.items[item].name),
          name: this.req.items[item].name,
          size: this.req.items[item].size,
          modified: this.req.items[item].modified,
          overwrite: false,
          rename: this.$route.path === this.dest,
        });
      }

      const action = async (overwrite, rename) => {
        buttons.loading("copy");

        await api
          .copy(items, overwrite, rename)
          .then(() => {
            buttons.success("copy");
            this.preselect = removePrefix(items[0].to);

            if (this.$route.path === this.dest) {
              this.reload = true;

              return;
            }

            if (this.user.redirectAfterCopyMove)
              this.$router.push({ path: this.dest });
          })
          .catch((e) => {
            buttons.done("copy");
            this.$showError(e);
          });
      };

      const dstItems = (await api.fetch(this.dest)).items;
      const conflict = upload.checkConflict(items, dstItems);

      if (conflict.length > 0) {
        this.showHover({
          prompt: "resolve-conflict",
          props: {
            conflict: conflict,
          },
          confirm: (event, result) => {
            event.preventDefault();
            this.closeHovers();
            for (let i = result.length - 1; i >= 0; i--) {
              const item = result[i];
              if (item.checked.length == 2) {
                items[item.index].rename = true;
              } else if (
                item.checked.length == 1 &&
                item.checked[0] == "origin"
              ) {
                items[item.index].overwrite = true;
              } else {
                items.splice(item.index, 1);
              }
            }
            if (items.length > 0) {
              action();
            }
          },
        });

        return;
      }

      action(false, false);
    },
  },
};
</script>
