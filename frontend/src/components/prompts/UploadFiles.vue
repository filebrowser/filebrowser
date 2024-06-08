<template>
  <div
    v-if="filesInUploadCount > 0"
    class="upload-files"
    v-bind:class="{ closed: !open }"
  >
    <div class="card floating">
      <div class="card-title">
        <h2>{{ $t("prompts.uploadFiles", { files: filesInUploadCount }) }}</h2>
        <div class="upload-info">
          <div class="upload-speed">{{ uploadSpeed.toFixed(2) }} MB/s</div>
          <div class="upload-eta">{{ formattedETA }} remaining</div>
          <div class="upload-percentage">
            {{ getProgressDecimal }}% Completed
          </div>
          <div class="upload-fraction">
            {{ getTotalProgressBytes }} / {{ getTotalSize }}
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
          v-for="file in filesInUpload"
          :key="file.id"
          :data-dir="file.isDir"
          :data-type="file.type"
          :aria-label="file.name"
        >
          <div class="file-name">
            <i class="material-icons"></i> {{ file.name }}
          </div>
          <div class="file-progress">
            <div v-bind:style="{ width: file.progress + '%' }"></div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { mapState, mapWritableState, mapActions } from "pinia";
import { useUploadStore } from "@/stores/upload";
import { useFileStore } from "@/stores/file";
import { abortAllUploads } from "@/api/tus";
import buttons from "@/utils/buttons";

export default {
  name: "uploadFiles",
  data: function () {
    return {
      open: false,
    };
  },
  computed: {
    ...mapState(useUploadStore, [
      "filesInUpload",
      "filesInUploadCount",
      "uploadSpeed",
      "getETA",
      "getProgress",
      "getProgressDecimal",
      "getTotalProgressBytes",
      "getTotalSize",
    ]),
    ...mapWritableState(useFileStore, ["reload"]),
    formattedETA() {
      if (!this.getETA || this.getETA === Infinity) {
        return "--:--:--";
      }

      let totalSeconds = this.getETA;
      const hours = Math.floor(totalSeconds / 3600);
      totalSeconds %= 3600;
      const minutes = Math.floor(totalSeconds / 60);
      const seconds = Math.round(totalSeconds % 60);

      return `${hours.toString().padStart(2, "0")}:${minutes
        .toString()
        .padStart(2, "0")}:${seconds.toString().padStart(2, "0")}`;
    },
  },
  methods: {
    ...mapActions(useUploadStore, ["reset"]), // Mapping reset action from upload store
    toggle: function () {
      this.open = !this.open;
    },
    abortAll() {
      if (confirm(this.$t("upload.abortUpload"))) {
        abortAllUploads();
        buttons.done("upload");
        this.open = false;
        this.reset(); // Resetting the upload store state
        this.reload = true; // Trigger reload in the file store
      }
    },
  },
};
</script>

<style scoped>
.upload-info {
  min-width: 19ch;
  width: auto;
  text-align: left;
}
</style>
