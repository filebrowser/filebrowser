<template>
  <div>
    <header-bar
      v-if="error || fileStore.req?.type === null"
      showMenu
      showLogo
    />

    <breadcrumbs base="/files" />
    <errors v-if="error" :errorCode="error.status" />
    <component v-else-if="currentView" :is="currentView"></component>
    <div v-else-if="currentView !== null">
      <h2 class="message delayed">
        <div class="spinner">
          <div class="bounce1"></div>
          <div class="bounce2"></div>
          <div class="bounce3"></div>
        </div>
        <span>{{ t("files.loading") }}</span>
      </h2>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  computed,
  defineAsyncComponent,
  onBeforeUnmount,
  onMounted,
  onUnmounted,
  ref,
  watch,
} from "vue";
import { files as api } from "@/api";
import { storeToRefs } from "pinia";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { useUploadStore } from "@/stores/upload";

import HeaderBar from "@/components/header/HeaderBar.vue";
import Breadcrumbs from "@/components/Breadcrumbs.vue";
import Errors from "@/views/Errors.vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";
import FileListing from "@/views/files/FileListing.vue";
import { StatusError } from "@/api/utils";
import { name } from "../utils/constants";

const Editor = defineAsyncComponent(() => import("@/views/files/Editor.vue"));
const Preview = defineAsyncComponent(() => import("@/views/files/Preview.vue"));

const layoutStore = useLayoutStore();
const fileStore = useFileStore();
const uploadStore = useUploadStore();

const { reload } = storeToRefs(fileStore);
const { error: uploadError } = storeToRefs(uploadStore);

const route = useRoute();

const { t } = useI18n({});

const clean = (path: string) => {
  return path.endsWith("/") ? path.slice(0, -1) : path;
};

const error = ref<StatusError | null>(null);

const currentView = computed(() => {
  if (fileStore.req?.type === undefined) {
    return null;
  }

  if (fileStore.req.isDir) {
    return FileListing;
  } else if (
    fileStore.req.type === "text" ||
    fileStore.req.type === "textImmutable"
  ) {
    return Editor;
  } else {
    return Preview;
  }
});

// Define hooks
onMounted(() => {
  fetchData();
  fileStore.isFiles = true;
  window.addEventListener("keydown", keyEvent);
});

onBeforeUnmount(() => {
  window.removeEventListener("keydown", keyEvent);
});

onUnmounted(() => {
  fileStore.isFiles = false;
  if (layoutStore.showShell) {
    layoutStore.toggleShell();
  }
  fileStore.updateRequest(null);
});

watch(route, (to, from) => {
  if (from.path.endsWith("/")) {
    window.sessionStorage.setItem(
      "listFrozen",
      (!to.path.endsWith("/")).toString()
    );
  } else if (to.path.endsWith("/")) {
    fileStore.updateRequest(null);
  }
  fetchData();
});
watch(reload, (newValue) => {
  newValue && fetchData();
});
watch(uploadError, (newValue) => {
  newValue && layoutStore.showError();
});

// Define functions

const fetchData = async () => {
  // Reset view information.
  fileStore.reload = false;
  fileStore.selected = [];
  fileStore.multiple = false;
  layoutStore.closeHovers();

  // Set loading to true and reset the error.
  if (
    window.sessionStorage.getItem("listFrozen") !== "true" &&
    window.sessionStorage.getItem("modified") !== "true"
  ) {
    layoutStore.loading = true;
  }
  error.value = null;

  let url = route.path;
  if (url === "") url = "/";
  if (url[0] !== "/") url = "/" + url;
  try {
    const res = await api.fetch(url);

    if (clean(res.path) !== clean(`/${[...route.params.path].join("/")}`)) {
      throw new Error("Data Mismatch!");
    }

    fileStore.updateRequest(res);
    document.title = `${res.name || t("sidebar.myFiles")} - ${t("files.files")} - ${name}`;
  } catch (err) {
    if (err instanceof Error) {
      error.value = err;
    }
  } finally {
    layoutStore.loading = false;
  }
};
const keyEvent = (event: KeyboardEvent) => {
  if (event.key === "F1") {
    event.preventDefault();
    layoutStore.showHover("help");
  }
};
</script>
