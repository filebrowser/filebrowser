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
          <div class="upload-percentage">{{ uploadPercentage.toFixed(2) }}% Completed</div>
          <div class="upload-size">
            {{ formatFileSize(totalUploadedSize) }} /
            {{ formatFileSize(totalFileSize) }}
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
import { mapGetters, mapMutations } from "vuex";
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
    ...mapGetters([
      "filesInUpload",
      "filesInUploadCount",
      "uploadSpeed",
      "eta",
      "uploadPercentage",
      "totalUploadedSize",
      "totalFileSize",
    ]),
    ...mapMutations(["resetUpload"]),
    formattedETA() {
      if (!this.eta || this.eta === Infinity) {
        return "--:--:--";
      }

      let totalSeconds = this.eta;
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
    toggle: function () {
      this.open = !this.open;
    },
    abortAll() {
      if (confirm(this.$t("upload.abortUpload"))) {
        abortAllUploads();
        buttons.done("upload");
        this.open = false;
        this.$store.commit("resetUpload");
        this.$store.commit("setReload", true);
      }
    },
    formatFileSize(size) {
      if (size < 1024) {
        return size + ' B';
      } else if (size < 1024 * 1024) {
        return (size / 1024).toFixed(2) + ' KB';
      } else if (size < 1024 * 1024 * 1024) {
        return (size / 1024 / 1024).toFixed(2) + ' MB';
      } else {
        return (size / 1024 / 1024 / 1024).toFixed(2) + ' GB';
      }
    },
  },
};
</script>
