<template>
  <div>
    <div v-if="uploadStore.totalBytes" class="progress">
      <div
        v-bind:style="{
          width: sentPercent + '%',
        }"
      ></div>
    </div>
    <sidebar v-if="!isCollaboraOfficeRoute"></sidebar>
    <main :class="{ 'office-fullscreen-main': isCollaboraOfficeRoute }">
      <router-view></router-view>
      <shell
        v-if="
          !isCollaboraOfficeRoute &&
          enableExec &&
          authStore.isLoggedIn &&
          authStore.user?.perm.execute
        "
      />
    </main>
    <prompts></prompts>
    <upload-files></upload-files>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import { useFileStore } from "@/stores/file";
import { useUploadStore } from "@/stores/upload";
import Sidebar from "@/components/Sidebar.vue";
import Prompts from "@/components/prompts/Prompts.vue";
import Shell from "@/components/Shell.vue";
import UploadFiles from "@/components/prompts/UploadFiles.vue";
import { enableExec } from "@/utils/constants";
import { computed, watch } from "vue";
import { useRoute } from "vue-router";

const layoutStore = useLayoutStore();
const authStore = useAuthStore();
const fileStore = useFileStore();
const uploadStore = useUploadStore();
const route = useRoute();

const sentPercent = computed(() =>
  ((uploadStore.sentBytes / uploadStore.totalBytes) * 100).toFixed(2)
);

const isCollaboraOfficeRoute = computed(
  () => route.path.startsWith("/files") && route.query.office === "true"
);

watch(route, () => {
  fileStore.selected = [];
  fileStore.multiple = false;
  if (layoutStore.currentPromptName !== "success") {
    layoutStore.closeHovers();
  }
});
</script>


<style scoped>
main.office-fullscreen-main {
  margin: 0;
  width: 100%;
}
</style>
