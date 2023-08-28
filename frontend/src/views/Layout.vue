<template>
  <div>
    <div v-if="progress" class="progress">
      <div v-bind:style="{ width: this.progress + '%' }"></div>
    </div>
    <sidebar></sidebar>
    <main>
      <router-view></router-view>
      <shell v-if="isExecEnabled && isLoggedIn && user.perm.execute" />
    </main>
    <prompts></prompts>
    <upload-files></upload-files>
  </div>
</template>

<script>
import { mapActions, mapState, mapWritableState } from "pinia";
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import { useFileStore } from "@/stores/file";
import Sidebar from "@/components/Sidebar.vue";
import Prompts from "@/components/prompts/Prompts.vue";
import Shell from "@/components/Shell.vue";
import UploadFiles from "@/components/prompts/UploadFiles.vue";
import { enableExec } from "@/utils/constants";

export default {
  name: "layout",
  components: {
    Sidebar,
    Prompts,
    Shell,
    UploadFiles,
  },
  computed: {
    ...mapState(useAuthStore, ["isLoggedIn", "user"]),
    ...mapState(useLayoutStore, ["progress", "show"]),
    ...mapWritableState(useFileStore, ["selected", "multiple"]),
    isExecEnabled: () => enableExec,
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers"]),
  },
  watch: {
    $route: function () {
      this.selected = [];
      this.multiple = false;
      if (this.show !== "success") {
        this.closeHovers();
      }
    },
  },
};
</script>
@/stores/auth@/stores/layout@/stores/file
