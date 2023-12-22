<template>
  <div>
    <div v-if="progress" class="progress">
      <div v-bind:style="{ width: this.progress + '%' }"></div>
    </div>
    <sidebar></sidebar>
    <main>
      <router-view></router-view>
      <shell v-if="isExecEnabled && isLogged && user.perm.execute" />
    </main>
    <prompts></prompts>
    <context-menu v-if="isVisibleContext"></context-menu>

    <upload-files></upload-files>
  </div>
</template>

<script>
import { mapState, mapGetters } from "vuex";
import Sidebar from "@/components/Sidebar.vue";
import Prompts from "@/components/prompts/Prompts.vue";
import ContextMenu from "@/components/files/ContextMenu.vue";
import Shell from "@/components/Shell.vue";
import UploadFiles from "../components/prompts/UploadFiles.vue";
import { enableExec } from "@/utils/constants";

export default {
  name: "layout",
  components: {
    Sidebar,
    Prompts,
    ContextMenu,
    Shell,
    UploadFiles,
  },
  computed: {
    ...mapGetters(["isLogged", "progress", "currentPrompt", "isVisibleContext"]),
    ...mapState(["user"]),
    isExecEnabled: () => enableExec,
  },
  watch: {
    $route: function () {
      this.$store.commit("hideContextMenu");
      this.$store.commit("resetSelected");
      this.$store.commit("multiple", false);
      if (this.currentPrompt?.prompt !== "success")
        this.$store.commit("closeHovers");
    },
  },
};
</script>
