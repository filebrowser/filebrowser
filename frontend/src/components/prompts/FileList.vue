<template>
  <div>
    <ul class="file-list">
      <li
        @click="itemClick"
        @touchstart="touchstart"
        @dblclick="next"
        role="button"
        tabindex="0"
        :aria-label="item.name"
        :aria-selected="selected == item.url"
        :key="item.name"
        v-for="item in items"
        :data-url="item.url"
      >
        {{ item.name }}
      </li>
    </ul>

    <p>
      {{ $t("prompts.currentlyNavigating") }} <code>{{ nav }}</code
      >.
    </p>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, inject, onMounted, onUnmounted } from "vue";
import { storeToRefs } from "pinia";
import { useRoute } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";

import url from "@/utils/url";
import { files } from "@/api";
import { StatusError } from "@/api/utils.js";

const props = defineProps<{
  exclude?: string[];
}>();

const emit = defineEmits<{
  "update:selected": [value: string];
}>();

const route = useRoute();
const $showError = inject<(error: unknown) => void>("$showError");

const authStore = useAuthStore();
const fileStore = useFileStore();
const layoutStore = useLayoutStore();

const { user } = storeToRefs(authStore);
const { req } = storeToRefs(fileStore);
const { showHover } = layoutStore;

const items = ref<Array<{ name: string; url: string }>>([]);
const touches = ref({
  id: "",
  count: 0,
});
const selected = ref<string | null>(null);
const current = ref(window.location.pathname);
const nextAbortController = ref(new AbortController());

const nav = computed(() => {
  return decodeURIComponent(current.value);
});

const abortOngoingNext = () => {
  nextAbortController.value.abort();
};

const fillOptions = (reqData: any) => {
  // Sets the current path and resets
  // the current items.
  current.value = reqData.url;
  items.value = [];

  emit("update:selected", current.value);

  // If the path isn't the root path,
  // show a button to navigate to the previous
  // directory.
  if (reqData.url !== "/files/") {
    items.value.push({
      name: "..",
      url: url.removeLastDir(reqData.url) + "/",
    });
  }

  // If this folder is empty, finish here.
  if (reqData.items === null) return;

  // Otherwise we add every directory to the
  // move options.
  for (const item of reqData.items) {
    if (!item.isDir) continue;
    if (props.exclude?.includes(item.url)) continue;

    items.value.push({
      name: item.name,
      url: item.url,
    });
  }
};

const next = (event: Event) => {
  // Retrieves the URL of the directory the user
  // just clicked in and fill the options with its
  // content.
  const uri = (event.currentTarget as HTMLElement).dataset.url!;
  abortOngoingNext();
  nextAbortController.value = new AbortController();
  files
    .fetch(uri, nextAbortController.value.signal)
    .then(fillOptions)
    .catch((e) => {
      if (e instanceof StatusError && e.is_canceled) {
        return;
      }
      $showError?.(e);
    });
};

const touchstart = (event: Event) => {
  const urlValue = (event.currentTarget as HTMLElement).dataset.url!;

  // In 300 milliseconds, we shall reset the count.
  setTimeout(() => {
    touches.value.count = 0;
  }, 300);

  // If the element the user is touching
  // is different from the last one he touched,
  // reset the count.
  if (touches.value.id !== urlValue) {
    touches.value.id = urlValue;
    touches.value.count = 1;
    return;
  }

  touches.value.count++;

  // If there is more than one touch already,
  // open the next screen.
  if (touches.value.count > 1) {
    next(event);
  }
};

const itemClick = (event: Event) => {
  if (user.value?.singleClick) next(event);
  else select(event);
};

const select = (event: Event) => {
  const urlValue = (event.currentTarget as HTMLElement).dataset.url!;
  // If the element is already selected, unselect it.
  if (selected.value === urlValue) {
    selected.value = null;
    emit("update:selected", current.value);
    return;
  }

  // Otherwise select the element.
  selected.value = urlValue;
  emit("update:selected", selected.value);
};

const createDir = async () => {
  showHover({
    prompt: "newDir",
    action: undefined,
    confirm: undefined,
    props: {
      redirect: false,
      base: current.value === route.path ? null : current.value,
    },
  });
};

onMounted(() => {
  if (req.value) {
    fillOptions(req.value);
  }
});

onUnmounted(() => {
  abortOngoingNext();
});

defineExpose({
  createDir,
});
</script>
