<template>
  <div>
    <component ref="currentComponent" :is="currentComponent"></component>
    <div v-show="showOverlay" @click="closeHovers" class="overlay"></div>
  </div>
</template>

<script>
import { mapActions, mapState } from "pinia";
import { useLayoutStore } from "@/stores/layout";

import Help from "./Help.vue";
import Info from "./Info.vue";
import Delete from "./Delete.vue";
import Rename from "./Rename.vue";
import Download from "./Download.vue";
import Move from "./Move.vue";
import Copy from "./Copy.vue";
import NewFile from "./NewFile.vue";
import NewDir from "./NewDir.vue";
import Replace from "./Replace.vue";
import ReplaceRename from "./ReplaceRename.vue";
import Share from "./Share.vue";
import Upload from "./Upload.vue";
import ShareDelete from "./ShareDelete.vue";
import buttons from "@/utils/buttons";

export default {
  name: "prompts",
  components: {
    Info,
    Delete,
    Rename,
    Download,
    Move,
    Copy,
    Share,
    NewFile,
    NewDir,
    Help,
    Replace,
    ReplaceRename,
    Upload,
    ShareDelete,
  },
  data: function () {
    return {
      pluginData: {
        buttons,
      },
    };
  },
  created() {
    window.addEventListener("keydown", (event) => {
      if (this.show == null) return;

      let prompt = this.$refs.currentComponent;

      // Esc!
      if (event.keyCode === 27) {
        event.stopImmediatePropagation();
        this.closeHovers();
      }

      // Enter
      if (event.keyCode == 13) {
        switch (this.show) {
          case "delete":
            prompt.submit();
            break;
          case "copy":
            prompt.copy(event);
            break;
          case "move":
            prompt.move(event);
            break;
          case "replace":
            prompt.showConfirm(event);
            break;
        }
      }
    });
  },
  computed: {
    ...mapState(useLayoutStore, ["show", "showConfirm"]),
    currentComponent: function () {
      const matched =
        [
          "info",
          "help",
          "delete",
          "rename",
          "move",
          "copy",
          "newFile",
          "newDir",
          "download",
          "replace",
          "replace-rename",
          "share",
          "upload",
          "share-delete",
        ].indexOf(this.show) >= 0;

      return (matched && this.show) || null;
    },
    showOverlay: function () {
      return (
        this.show !== null && this.show !== "search" && this.show !== "more"
      );
    },
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers"]),
  },
};
</script>
@/stores/layout
