<template>
  <div id="context-menu" ref="contextMenuDiv" class="card" :style="menuStyle">
    <p>
      <action icon="info" :label="t('buttons.info')" show="info" />
    </p>
    <p v-if="options.share">
      <action icon="share" :label="t('buttons.share')" show="share" />
    </p>
    <p v-if="options.edit">
      <action icon="mode_edit" :label="t('buttons.edit')" @action="openFile" />
    </p>
    <p v-if="options.rename">
      <action
        icon="drive_file_rename_outline"
        :label="t('buttons.rename')"
        show="rename"
      />
    </p>
    <p v-if="options.copy">
      <action
        id="copy-button"
        icon="content_copy"
        :label="t('buttons.copyFile')"
        show="copy"
      />
    </p>
    <p v-if="options.move">
      <action
        id="move-button"
        icon="forward"
        :label="t('buttons.moveFile')"
        show="move"
      />
    </p>
    <p v-if="options.permissions">
      <action
        id="permissions-button"
        icon="lock"
        :label="t('buttons.permissions')"
        show="permissions"
      />
    </p>
    <p v-if="options.archive">
      <action
        id="archive-button"
        icon="archive"
        :label="t('buttons.archive')"
        show="archive"
      />
    </p>
    <p v-if="options.unarchive">
      <action
        id="unarchive-button"
        icon="unarchive"
        :label="t('buttons.unarchive')"
        show="unarchive"
      />
    </p>
    <p v-if="options.download">
      <action
        icon="file_download"
        :label="t('buttons.download')"
        @action="download"
        :counter="fileStore.selectedCount"
      />
    </p>
    <p v-if="options.delete">
      <action
        id="delete-button"
        icon="delete"
        :label="t('buttons.delete')"
        show="delete"
      />
    </p>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, type CSSProperties } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { useContextMenuStore } from "@/stores/contextMenu";
import { files as api } from "@/api";
import { useI18n } from "vue-i18n";
import Action from "../header/Action.vue";

const authStore = useAuthStore();
const fileStore = useFileStore();
const layoutStore = useLayoutStore();
const contextMenuStore = useContextMenuStore();

const route = useRoute();
const router = useRouter();

const { t } = useI18n();

const contextMenuDiv = ref<HTMLDivElement | null>(null);

const menuStyle = computed((): CSSProperties => {
  if (contextMenuStore.position === null) {
    return { left: "0px", right: "0px" };
  }

  let style: CSSProperties = {
    left: contextMenuStore.position.x + "px",
    top: contextMenuStore.position.y + "px",
  };

  if (window.innerWidth - contextMenuStore.position.x < 150) {
    style.transform = "translateX(calc(-100% - 3px))";
  }

  return style;
});

const options = computed(() => {
  return {
    download: authStore.user?.perm.download ?? false,
    delete: fileStore.selectedCount > 0 && authStore.user?.perm.delete,
    edit:
      fileStore.selectedCount === 1 &&
      (fileStore.req?.items[fileStore.selected[0]].type === "text" ||
        fileStore.req?.items[fileStore.selected[0]].type === "textImmutable"),
    rename: fileStore.selectedCount === 1 && authStore.user?.perm.rename,
    share: fileStore.selectedCount === 1 && authStore.user?.perm.share,
    move: fileStore.selectedCount > 0 && authStore.user?.perm.rename,
    copy: fileStore.selectedCount > 0 && authStore.user?.perm.create,
    permissions: fileStore.selectedCount === 1 && authStore.user?.perm.modify,
    archive: fileStore.selectedCount > 0 && authStore.user?.perm.create,
    unarchive:
      fileStore.selectedCount === 1 &&
      fileStore.onlyArchivesSelected &&
      authStore.user?.perm.create,
  };
});

const windowClick = (event: MouseEvent) => {
  if (!contextMenuDiv.value?.contains(event.target as Node)) {
    contextMenuStore.hide();
  }
};

const openFile = () => {
  let path = fileStore.req?.items[fileStore.selected[0]].url;
  if (path) {
    router.push({ path: path });
  }
};

const download = () => {
  if (fileStore.req === null) return;

  if (
    fileStore.selectedCount === 1 &&
    !fileStore.req.items[fileStore.selected[0]].isDir
  ) {
    api.download(null, fileStore.req.items[fileStore.selected[0]].url);
    return;
  }

  layoutStore.showHover({
    prompt: "download",
    confirm: (format: any) => {
      layoutStore.closeHovers();

      let files = [];

      if (fileStore.selectedCount > 0 && fileStore.req !== null) {
        for (let i of fileStore.selected) {
          files.push(fileStore.req.items[i].url);
        }
      } else {
        files.push(route.path);
      }

      api.download(format, ...files);
    },
  });
};

onMounted(() => {
  window.addEventListener("mousedown", windowClick);
});

onBeforeUnmount(() => {
  window.removeEventListener("mousedown", windowClick);
});
</script>
