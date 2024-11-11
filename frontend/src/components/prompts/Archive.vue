<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ t("prompts.archive") }}</h2>
    </div>

    <div class="card-content">
      <p>{{ t("prompts.archiveMessage") }}</p>
      <input
        class="input input--block"
        v-focus
        type="text"
        v-model.trim="name"
        :disabled="loading"
        required
      />

      <button
        v-for="(ext, format) in formats"
        :key="format"
        :disabled="loading"
        class="button button--block"
        @click="archive(format)"
      >
        <i
          v-if="loading && format === loadingFormat"
          class="material-icons spin"
        >
          autorenew
        </i>
        <span v-else>{{ ext }}</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { inject, ref } from "vue";
import { useI18n } from "vue-i18n";
import { onMounted } from "vue";
import { useRoute } from 'vue-router'
import { useFileStore } from "@/stores/file";
import { useQuotaStore } from "@/stores/quota";
import { useLayoutStore } from "@/stores/layout";
import { files as api } from "@/api";
import url from "@/utils/url";
import buttons from "@/utils/buttons";

const fileStore = useFileStore();
const quotaStore = useQuotaStore();
const layoutStore = useLayoutStore();

const route = useRoute();

const { t } = useI18n();

const $showError = inject<IToastError>("$showError")!;

const formats = {
  zip: "zip",
  tar: "tar",
  targz: "tar.gz",
  tarbz2: "tar.bz2",
  tarxz: "tar.xz",
  tarlz4: "tar.lz4",
  tarsz: "tar.sz",
};

const name = ref<string>("");
const loading = ref<boolean>(false);
const loadingFormat = ref<string>("");

const archive = async (format: string) => {
  let items: string[] = [];

  for (let i of fileStore.selected) {
    let item = fileStore.req?.items[i].name
    if (item) {
      items.push(item);
    }
  }

  let uri = fileStore.isFiles ? route.path : "/";

  if (!fileStore.isListing) {
    uri = url.removeLastDir(uri);
  }

  uri += "/archive";
  uri = uri.replace("//", "/");

  try {
    loading.value = true;
    loadingFormat.value = format;
    buttons.loading("archive");
    await api.archive(uri, name.value, format, ...items);
    layoutStore.closeHovers();
    fileStore.reload = true;
    quotaStore.fetchQuota(3000);
  } catch (e: any) {
    $showError(e);
  } finally {
    loading.value = false;
    loadingFormat.value = "";
    buttons.done("archive");
  }
};

onMounted(() => {
  if (fileStore.selected.length > 0) {
    name.value = fileStore.req?.items[fileStore.selected[0]].name || "";
  }
});
</script>
