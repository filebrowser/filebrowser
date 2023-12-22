<template>
  <div>
    <component
      v-if="showOverlay"
      :ref="currentPromptName"
      :is="currentPromptName"
      v-bind="currentPrompt.props"
    >
    </component>
    <div v-show="showOverlay" @click="resetPrompts" class="overlay"></div>
  </div>
</template>

<script>
import Help from "./Help.vue";
import Info from "./Info.vue";
import Delete from "./Delete.vue";
import Rename from "./Rename.vue";
import Download from "./Download.vue";
import Move from "./Move.vue";
import Archive from "./Archive.vue";
import Unarchive from "./Unarchive.vue";
import Permissions from "./Permissions.vue";
import Copy from "./Copy.vue";
import NewFile from "./NewFile.vue";
import NewDir from "./NewDir.vue";
import Replace from "./Replace.vue";
import ReplaceRename from "./ReplaceRename.vue";
import Share from "./Share.vue";
import Upload from "./Upload.vue";
import ShareDelete from "./ShareDelete.vue";
import Sidebar from "../Sidebar.vue";
import { mapGetters, mapState } from "vuex";
import buttons from "@/utils/buttons";

export default {
  name: "prompts",
  components: {
    Info,
    Delete,
    Rename,
    Download,
    Move,
    Archive,
    Unarchive,
    Permissions,
    Copy,
    Share,
    NewFile,
    NewDir,
    Help,
    Replace,
    ReplaceRename,
    Upload,
    ShareDelete,
    Sidebar
  },
  data: function () {
    return {
      pluginData: {
        buttons,
        store: this.$store,
        router: this.$router,
      },
    };
  },
  created() {
    window.addEventListener("keydown", (event) => {
      if (this.currentPrompt == null) return;

      let prompt = this.$refs.currentComponent;

      // Esc!
      if (event.keyCode === 27) {
        event.stopImmediatePropagation();
        this.$store.commit("closeHovers");
      }

      // Enter
      if (event.keyCode == 13) {
        switch (this.currentPrompt.prompt) {
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
    ...mapState(["plugins"]),
    ...mapGetters(["currentPrompt", "currentPromptName"]),
    showOverlay: function () {
      return (
        this.currentPrompt !== null &&
        this.currentPrompt.prompt !== "search" &&
        this.currentPrompt.prompt !== "more"
      );
    },
  },
  methods: {
    resetPrompts() {
      this.$store.commit("closeHovers");
    },
  },
};
</script>
