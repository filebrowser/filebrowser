<template>
  <div
    v-if="uploadStore.activeUploads.size > 0"
    class="upload-files"
    v-bind:class="{ closed: !open }"
  >
    <div class="card floating">
      <div class="card-title">
        <h2>
          {{
            $t("prompts.uploadFiles", { files: uploadStore.activeUploads.size })
          }}
        </h2>
        <div class="upload-info">
          <div class="upload-speed">{{ speed.toFixed(2) }} MB/s</div>
          <div class="upload-eta">{{ formattedETA }} remaining</div>
          <div class="upload-percentage">
            {{ uploadStore.getProgressDecimal }}% Completed
          </div>
          <div class="upload-fraction">
            {{ uploadStore.getTotalProgressBytes }} /
            {{ uploadStore.getTotalSize }}
          </div>
        </div>
        <button
          class="action"
          @click="abortAll"
          aria-label="Abort upload"
          title="Abort upload"
        >
          <i class="material-icons">{{ "cancel" }}</i>
        </button>
        <button
          class="action"
          @click="toggle"
          aria-label="Toggle file upload list"
          title="Toggle file upload list"
        >
          <i class="material-icons">{{
            open ? "keyboard_arrow_down" : "keyboard_arrow_up"
          }}</i>
        </button>
      </div>

      <div class="card-content file-icons">
        <div
          class="file"
          v-for="upload in uploadStore.activeUploads"
          :key="upload.path"
          :data-dir="upload.type === 'dir'"
          :data-type="upload.type"
          :aria-label="upload.name"
        >
          <div class="file-name">
            <i class="material-icons"></i> {{ upload.name }}
          </div>
          <div class="file-progress">
            <div
              v-bind:style="{
                width: (upload.sentBytes / upload.totalBytes) * 100 + '%',
              }"
            ></div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useFileStore } from "@/stores/file";
import { useUploadStore } from "@/stores/upload";
import { storeToRefs } from "pinia";
import { computed, ref, watch } from "vue";
import { abortAllUploads } from "@/api/tus";
import buttons from "@/utils/buttons";
import { useI18n } from "vue-i18n";

const { t } = useI18n({});

const open = ref<boolean>(false);
const speed = ref<number>(0);
const eta = ref<number>(Infinity);

const fileStore = useFileStore();
const uploadStore = useUploadStore();

const { sentBytes } = storeToRefs(uploadStore);

let lastSpeedUpdate: number = 0;
const recentSpeeds: number[] = [];

const calculateSpeed = (sentBytes: number, oldSentBytes: number) => {
  const elapsedTime = (Date.now() - (lastSpeedUpdate ?? 0)) / 1000;
  const bytesSinceLastUpdate = sentBytes - oldSentBytes;
  const currentSpeed = bytesSinceLastUpdate / (1024 * 1024) / elapsedTime;

  recentSpeeds.push(currentSpeed);
  if (recentSpeeds.length > 5) {
    recentSpeeds.shift();
  }

  const recentSpeedsAverage =
    recentSpeeds.reduce((acc, curr) => acc + curr) / recentSpeeds.length;

  speed.value = recentSpeedsAverage * 0.2 + speed.value * 0.8;
  lastSpeedUpdate = Date.now();

  calculateEta();
};

const calculateEta = () => {
  if (speed.value === 0) {
    eta.value = Infinity;

    return Infinity;
  }

  const remainingSize = uploadStore.totalBytes - uploadStore.sentBytes;
  const speedBytesPerSecond = speed.value * 1024 * 1024;

  eta.value = remainingSize / speedBytesPerSecond;
};

watch(sentBytes, calculateSpeed);

const formattedETA = computed(() => {
  if (!eta.value || eta.value === Infinity) {
    return "--:--:--";
  }

  let totalSeconds = eta.value;
  const hours = Math.floor(totalSeconds / 3600);
  totalSeconds %= 3600;
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = Math.round(totalSeconds % 60);

  return `${hours.toString().padStart(2, "0")}:${minutes
    .toString()
    .padStart(2, "0")}:${seconds.toString().padStart(2, "0")}`;
});

const toggle = () => {
  open.value = !open.value;
};

const abortAll = () => {
  if (confirm(t("upload.abortUpload"))) {
    abortAllUploads();
    buttons.done("upload");
    open.value = false;
    uploadStore.reset(); // Resetting the upload store state
    fileStore.reload = true; // Trigger reload in the file store
  }
};
</script>

<style scoped>
.upload-info {
  min-width: 19ch;
  width: auto;
  text-align: left;
}
</style>
