<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ t("prompts.newFile") }}</h2>
    </div>

    <div class="card-content">
      <p>{{ t("prompts.newFileMessage") }}</p>

      <input id="focus-prompt" class="input input--block" type="text" @keyup.enter="submit" v-model.trim="name" />
      <div class="path-container" ref="pathContainer">
        <template v-for="(item, index) in path" :key="index">
          /
          <span class="path-item">
            <span v-if="(index == path.length - 1) && item.includes('.')"
              class="material-icons">insert_drive_file</span>
            <span v-else class="material-icons">folder</span>
            {{ item }}
          </span>
        </template>

      </div>
    </div>

    <div class="card-action">
      <button class="button button--flat button--grey" @click="layoutStore.closeHovers"
        :aria-label="t('buttons.cancel')" :title="t('buttons.cancel')">
        {{ t("buttons.cancel") }}
      </button>
      <button class="button button--flat" @click="submit" :aria-label="t('buttons.create')"
        :title="t('buttons.create')">
        {{ t("buttons.create") }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { inject, ref, computed, watch, nextTick } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute, useRouter } from "vue-router";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";

import { files as api } from "@/api";
import url from "@/utils/url";

const $showError = inject<IToastError>("$showError")!;

const fileStore = useFileStore();
const layoutStore = useLayoutStore();

const pathContainer = ref<HTMLElement | null>(null);

const route = useRoute();
const router = useRouter();
const { t } = useI18n();

const name = ref<string>("");

const path = computed(() => {
  let basePath = fileStore.isFiles ? route.path : url.removeLastDir(route.path);
  basePath += name.value;

  return basePath
    .replace(/^\/[^\/]+/, '')
    .split('/')
    .filter(Boolean);
});

watch(path, () => {
  nextTick(() => {
    const lastItem = pathContainer.value?.lastElementChild;
    lastItem?.scrollIntoView({ behavior: 'auto', inline: 'end' });
  });
});


const submit = async (event: Event) => {
  event.preventDefault();
  if (name.value === "") return;

  // Build the path of the new directory.
  let uri = fileStore.isFiles ? route.path + "/" : "/";

  if (!fileStore.isListing) {
    uri = url.removeLastDir(uri) + "/";
  }

  uri += encodeURIComponent(name.value);
  uri = uri.replace("//", "/");

  try {
    await api.post(uri);
    router.push({ path: uri });
  } catch (e) {
    if (e instanceof Error) {
      $showError(e);
    }
  }

  layoutStore.closeHovers();
};
</script>
