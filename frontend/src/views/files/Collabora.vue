<template>
  <div id="collabora-container">
    <header-bar>
      <action icon="close" :label="t('buttons.close')" @action="close" />
      <span class="collabora-title">{{ fileStore.req?.name ?? title }}</span>

      <template #actions>
        <action
          icon="refresh"
          :label="t('buttons.reloadCollabora')"
          @action="loadEditor"
        />
        <action
          :icon="isNativeFullscreen ? 'fullscreen_exit' : 'fullscreen'"
          :label="isNativeFullscreen ? t('buttons.exitFullscreen') : t('buttons.fullscreen')"
          @action="toggleNativeFullscreen"
        />
        <action
          v-if="fileStore.req"
          icon="file_download"
          :label="t('buttons.download')"
          @action="download"
        />
      </template>
    </header-bar>

    <div class="loading delayed" v-if="layoutStore.loading">
      <div class="spinner">
        <div class="bounce1"></div>
        <div class="bounce2"></div>
        <div class="bounce3"></div>
      </div>
    </div>

    <iframe
      v-else-if="editorUrl"
      :key="editorUrl"
      :src="editorUrl"
      allow="clipboard-read; clipboard-write; fullscreen"
      referrerpolicy="no-referrer-when-downgrade"
    ></iframe>

    <div v-else class="info">
      <div class="title">
        <i class="material-icons">feedback</i>
        {{ errorMessage || "Collabora could not be opened." }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { inject, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";

import { collabora, files as api } from "@/api";
import url from "@/utils/url";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import HeaderBar from "@/components/header/HeaderBar.vue";
import Action from "@/components/header/Action.vue";

const fileStore = useFileStore();
const layoutStore = useLayoutStore();
const route = useRoute();
const router = useRouter();
const { t } = useI18n();

const $showError = inject<IToastError>("$showError")!;

const editorUrl = ref("");
const errorMessage = ref("");
const title = ref("Collabora");
const isNativeFullscreen = ref(false);

const loadEditor = async () => {
  if (!fileStore.req) return;

  editorUrl.value = "";
  errorMessage.value = "";
  layoutStore.loading = true;

  try {
    const response = await collabora.openPath(fileStore.req.path);
    title.value = response.name;
    editorUrl.value = response.url;
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : String(error);
    $showError(error as Error);
  } finally {
    layoutStore.loading = false;
  }
};

const close = () => {
  if (fileStore.req) {
    fileStore.preselect = fileStore.req.path;
  }

  const uri = url.removeLastDir(route.path) + "/";
  router.push({ path: uri });
};

const handleCollaboraMessage = (event: MessageEvent) => {
  if (!editorUrl.value) return;

  try {
    const editorOrigin = new URL(editorUrl.value).origin;
    if (event.origin !== editorOrigin) return;
  } catch {
    return;
  }

  let data: any = event.data;
  if (typeof data === "string") {
    try {
      data = JSON.parse(data);
    } catch {
      data = { MessageId: data };
    }
  }

  const messageId = data?.MessageId ?? data?.messageId ?? data?.type;
  if (["UI_Close", "UI_CloseClicked", "close", "Close"].includes(messageId)) {
    close();
  }
};

const download = () => {
  if (fileStore.req) {
    window.open(api.getDownloadURL(fileStore.req, false));
  }
};

const syncFullscreenState = () => {
  isNativeFullscreen.value = !!document.fullscreenElement;
};

const toggleNativeFullscreen = async () => {
  const container = document.getElementById("collabora-container");

  try {
    if (document.fullscreenElement) {
      await document.exitFullscreen();
    } else if (container?.requestFullscreen) {
      await container.requestFullscreen();
    }
  } catch (error) {
    $showError(error as Error);
  }
};

onMounted(() => {
  document.body.classList.add("collabora-open");
  document.addEventListener("fullscreenchange", syncFullscreenState);
  window.addEventListener("message", handleCollaboraMessage);
  loadEditor();
});

onBeforeUnmount(() => {
  document.body.classList.remove("collabora-open");
  document.removeEventListener("fullscreenchange", syncFullscreenState);
  window.removeEventListener("message", handleCollaboraMessage);
});
watch(
  () => fileStore.req?.path,
  () => loadEditor()
);
</script>

<style scoped>
:global(body.collabora-open) {
  overflow: hidden;
}

#collabora-container {
  position: fixed;
  inset: 0;
  z-index: 200000;
  display: flex;
  flex-direction: column;
  height: 100dvh;
  min-height: 100dvh;
  padding-top: 4em;
  box-sizing: border-box;
  background: var(--background);
}

#collabora-container :deep(header) {
  z-index: 200001;
}

#collabora-container iframe {
  border: 0;
  flex: 1 1 auto;
  height: calc(100dvh - 4em);
  width: 100%;
  background: white;
}

#collabora-container .loading,
#collabora-container .info {
  flex: 1 1 auto;
}

.collabora-title {
  flex: 1 1 auto;
  padding: 0 1em;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 600;
}
</style>
