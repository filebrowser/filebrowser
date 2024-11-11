<template>
  <div
    class="item"
    role="button"
    tabindex="0"
    :draggable="isDraggable"
    @dragstart="dragStart"
    @dragover="dragOver"
    @drop="drop"
    @click="itemClick"
    :data-dir="isDir"
    :data-type="type"
    :aria-label="name"
    :aria-selected="isSelected"
    :data-ext="getExtension(name).toLowerCase()"
    @contextmenu.prevent="contextMenu"
  >
    <div>
      <img
        v-if="!readOnly && type === 'image' && isThumbsEnabled"
        v-lazy="thumbnailUrl"
      />
      <i v-else class="material-icons"></i>
    </div>

    <div>
      <p v-if="isSymlink && link !== ''" class="name">
        {{ name }} → {{ link }}
      </p>
      <p v-else class="name">{{ name }}</p>

      <p class="size" :data-order="diskUsage?.size || humanSize() || '-1'">
        {{ usedDiskSize }}
      </p>

      <p class="modified">
        <time :datetime="modified">{{ humanTime() }}</time>
      </p>

      <p class="permissions">{{ permissions() }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { useContextMenuStore } from "@/stores/contextMenu";

import { eventPosition } from "@/utils/event";
import { enableThumbs } from "@/utils/constants";
import { filesize } from "@/utils";
import dayjs from "dayjs";
import { files as api } from "@/api";
import * as upload from "@/utils/upload";
import { computed, inject, ref, watch, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { storeToRefs } from "pinia";

const { t } = useI18n();

const touches = ref<number>(0);

const $showError = inject<IToastError>("$showError")!;
const router = useRouter();

const props = defineProps<{
  name: string;
  link: string;
  isDir: boolean;
  isSymlink: boolean;
  url: string;
  type: string;
  size: number;
  mode: number;
  modified: string;
  index: number;
  readOnly?: boolean;
  path?: string;
}>();

const authStore = useAuthStore();
const fileStore = useFileStore();
const layoutStore = useLayoutStore();
const contextMenuStore = useContextMenuStore();

const { diskUsages } = storeToRefs(fileStore);

const diskUsage = ref<DiskUsage | null>(null);

const usedDiskSize = computed((): string => {
  if (props.isDir) {
    if (!diskUsage.value) {
      return "-";
    }

    return (
      diskUsage.value.size +
      " " +
      t("prompts.inodeCount", { count: diskUsage.value.inodes })
    );
  }

  return humanSize();
});
const singleClick = computed(
  () => !props.readOnly && authStore.user?.singleClick
);
const isSelected = computed(
  () => fileStore.selected.indexOf(props.index) !== -1
);
const isDraggable = computed(
  () => !props.readOnly && authStore.user?.perm.rename
);

watch(
  diskUsages,
  () => {
    updateDiskUsage();
  },
  { deep: true }
);

const canDrop = computed(() => {
  if (!props.isDir || props.readOnly) return false;

  for (let i of fileStore.selected) {
    if (fileStore.req?.items[i].url === props.url) {
      return false;
    }
  }

  return true;
});

const thumbnailUrl = computed(() => {
  const file = {
    path: props.path,
    modified: props.modified,
  };

  return api.getPreviewURL(file as Resource, "thumb");
});

const isThumbsEnabled = computed(() => {
  return enableThumbs;
});

const humanSize = () => {
  return props.type == "invalid_link" ? "invalid link" : filesize(props.size);
};

const permissions = () => {
  let s = "";
  if (props.isSymlink) {
    s += "l";
  } else if (props.isDir) {
    s += "d";
  } else {
    s += "-";
  }
  s += (props.mode & 256) != 0 ? "r" : "-";
  s += (props.mode & 128) != 0 ? "w" : "-";
  s += (props.mode & 64) != 0 ? "x" : "-";
  s += (props.mode & 32) != 0 ? "r" : "-";
  s += (props.mode & 16) != 0 ? "w" : "-";
  s += (props.mode & 8) != 0 ? "x" : "-";
  s += (props.mode & 4) != 0 ? "r" : "-";
  s += (props.mode & 2) != 0 ? "w" : "-";
  s += (props.mode & 1) != 0 ? "x" : "-";
  return s;
};

const updateDiskUsage = () => {
  if (props.path) {
    diskUsage.value = fileStore.diskUsages.get(props.path) || null;
  }
};

const humanTime = () => {
  if (!props.readOnly && authStore.user?.dateFormat) {
    return dayjs(props.modified).format("L LT");
  }
  return dayjs(props.modified).fromNow();
};

const dragStart = () => {
  if (fileStore.selectedCount === 0) {
    fileStore.selected.push(props.index);
    return;
  }

  if (!isSelected.value) {
    fileStore.selected = [];
    fileStore.selected.push(props.index);
  }
};

const dragOver = (event: Event) => {
  if (!canDrop.value) return;

  event.preventDefault();
  let el = event.target as HTMLElement | null;
  if (el !== null) {
    for (let i = 0; i < 5; i++) {
      if (!el?.classList.contains("item")) {
        el = el?.parentElement ?? null;
      }
    }

    if (el !== null) el.style.opacity = "1";
  }
};

const drop = async (event: Event) => {
  if (!canDrop.value) return;
  event.preventDefault();

  if (fileStore.selectedCount === 0) return;

  let el = event.target as HTMLElement | null;
  for (let i = 0; i < 5; i++) {
    if (el !== null && !el.classList.contains("item")) {
      el = el.parentElement;
    }
  }

  let items: any[] = [];

  for (let i of fileStore.selected) {
    if (fileStore.req) {
      items.push({
        from: fileStore.req?.items[i].url,
        to: props.url + encodeURIComponent(fileStore.req?.items[i].name),
        name: fileStore.req?.items[i].name,
      });
    }
  }

  // Get url from ListingItem instance
  if (el === null) {
    return;
  }
  let path = el.__vue__.url;
  let baseItems = (await api.fetch(path)).items;

  let action = (overwrite: boolean, rename: boolean) => {
    api
      .move(items, overwrite, rename)
      .then(() => {
        fileStore.reload = true;
      })
      .catch($showError);
  };

  let conflict = upload.checkConflict(items, baseItems);

  let overwrite = false;
  let rename = false;

  if (conflict) {
    layoutStore.showHover({
      prompt: "replace-rename",
      confirm: (event: Event, option: any) => {
        overwrite = option == "overwrite";
        rename = option == "rename";

        event.preventDefault();
        layoutStore.closeHovers();
        action(overwrite, rename);
      },
    });

    return;
  }

  action(overwrite, rename);
};

const itemClick = (event: Event | KeyboardEvent) => {
  if (
    singleClick.value &&
    !(event as KeyboardEvent).ctrlKey &&
    !(event as KeyboardEvent).metaKey &&
    !(event as KeyboardEvent).shiftKey &&
    !fileStore.multiple
  )
    open();
  else click(event);
};

const click = (event: Event | KeyboardEvent) => {
  if (!singleClick.value && fileStore.selectedCount !== 0)
    event.preventDefault();

  setTimeout(() => {
    touches.value = 0;
  }, 300);

  touches.value++;
  if (touches.value > 1) {
    open();
  }

  if (fileStore.selected.indexOf(props.index) !== -1) {
    fileStore.removeSelected(props.index);
    return;
  }

  if ((event as KeyboardEvent).shiftKey && fileStore.selected.length > 0) {
    let fi = 0;
    let la = 0;

    if (props.index > fileStore.selected[0]) {
      fi = fileStore.selected[0] + 1;
      la = props.index;
    } else {
      fi = props.index;
      la = fileStore.selected[0] - 1;
    }

    for (; fi <= la; fi++) {
      if (fileStore.selected.indexOf(fi) == -1) {
        fileStore.selected.push(fi);
      }
    }

    return;
  }

  if (
    !singleClick.value &&
    !(event as KeyboardEvent).ctrlKey &&
    !(event as KeyboardEvent).metaKey &&
    !fileStore.multiple
  ) {
    fileStore.selected = [];
  }
  fileStore.selected.push(props.index);
};

const open = () => {
  router.push({ path: props.url });
};

const getExtension = (fileName: string): string => {
  const lastDotIndex = fileName.lastIndexOf(".");
  if (lastDotIndex === -1) {
    return fileName;
  }
  return fileName.substring(lastDotIndex);
};

const contextMenu = (event: MouseEvent) => {
  contextMenuStore.hide();

  if (fileStore.selected.indexOf(props.index) === -1) {
    fileStore.selected = [props.index];
  }

  let pos = eventPosition(event);

  contextMenuStore.show(pos.x + 2, pos.y);
};

onMounted(() => {
  updateDiskUsage();
});
</script>
