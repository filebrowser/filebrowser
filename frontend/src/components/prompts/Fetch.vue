<template>
  <div class="card floating" id="fetch">
    <div class="card-title">
      <h2>{{ t("prompts.fetch") }}</h2>
    </div>

    <div class="card-content">
      <p>{{ t("prompts.fetchUrl") }}</p>
      <input
        id="focus-prompt"
        class="input input--block"
        type="text"
        @keyup.enter="submit"
        v-model.trim="fetchUrl"
        tabindex="1"
        :disabled="isDownloading"
      />
    </div>

    <div class="card-content">
      <p>{{ t("prompts.fetchSaveName") }}</p>
      <input
        class="input input--block"
        type="text"
        @keyup.enter="submit"
        v-model.trim="saveName"
        tabindex="2"
        :disabled="isDownloading"
      />
    </div>

    <div class="card-action">
      <button
        class="button button--flat button--grey"
        @click="layoutStore.closeHovers"
        :aria-label="t('buttons.cancel')"
        :title="t('buttons.cancel')"
        tabindex="4"
      >
        {{ t("buttons.cancel") }}
      </button>
      <button
        class="button button--flat"
        :aria-label="t('buttons.create')"
        :title="t('buttons.create')"
        @click="submit"
        tabindex="3"
        :disabled="isDownloading"
      >
        {{ t("buttons.create") }}
      </button>
    </div>
    <div v-if="progress > 0" class="material-progress-container">
      <div
        class="material-progress-bar"
        id="downloadProgress"
        :style="{ width: Math.floor(progress * 100) + '%' }"
      ></div>
      <div class="material-progress-label">
        {{ Math.floor((progress * 100)) + "%" }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useLayoutStore } from "@/stores/layout.ts";
import { useFileStore } from "@/stores/file.ts";
import { useI18n } from "vue-i18n";
import { computed, inject, onMounted, ref, watch } from "vue";
import url from "@/utils/url.ts";
import { fetcher as api } from "@/api";
import { useRoute } from "vue-router";

const $showError = inject<IToastError>("$showError")!;

const layoutStore = useLayoutStore();
const fileStore = useFileStore();
const route = useRoute();

const { t } = useI18n();

const fetchUrl = ref<string>("");
const saveName = ref<string>("");
const taskID = ref<string>("");
const progress = ref<number>(0);

const isDownloading = computed(() => {
  return taskID.value !== "" && progress.value < 1 && progress.value > 0;
});

watch(fetchUrl, (value) => {
  try {
    if (saveName.value !== "") return;
    if (!(value.startsWith("http://") || value.startsWith("https://"))) return;
    saveName.value = new URL(value).pathname.split("/").pop() ?? "";
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
  } catch (e) {}
});

const submit = async (event: Event) => {
  event.preventDefault();
  if (
    !(
      fetchUrl.value.startsWith("http://") ||
      fetchUrl.value.startsWith("https://")
    )
  )
    return;
  if (saveName.value === "") return;

  // Build the path of the new directory.
  let uri = fileStore.isFiles ? route.path + "/" : "/";

  if (!fileStore.isListing) {
    uri = url.removeLastDir(uri) + "/";
  }

  try {
    const createdTaskID = await api.fetchUrlFile(
      uri,
      saveName.value,
      fetchUrl.value,
      {}
    );
    if (createdTaskID) {
      taskID.value = createdTaskID;
    }
  } catch (e) {
    if (e instanceof Error) {
      $showError(e);
    }
  }
};

onMounted(() => {
  setInterval(async () => {
    if (!taskID.value) return;
    const task = await api.queryDownloadTask(taskID.value);
    if (!task) return;
    console.log("fetch task info", task);
    progress.value = task.progress;
    if (task.progress >= 1) {
      taskID.value = "";
    }
  }, 1000);
});
</script>

<style scoped></style>
