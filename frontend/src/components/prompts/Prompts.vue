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
import Help from "./Help";
import Info from "./Info";
import Delete from "./Delete";
import Rename from "./Rename";
import Download from "./Download";
import Move from "./Move";
import Copy from "./Copy";
import NewFile from "./NewFile";
import NewDir from "./NewDir";
import Replace from "./Replace";
import ReplaceRename from "./ReplaceRename";
import Share from "./Share";
import Upload from "./Upload";
import ShareDelete from "./ShareDelete";
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
