<template>
  <div
    v-if="filesInUploadCount > 0"
    class="upload-files"
    v-bind:class="{ closed: !open }"
  >
    <div class="card floating">
      <div class="card-title">
        <h2>{{ $t("prompts.uploadFiles", { files: filesInUploadCount }) }}</h2>

        <button
          class="action"
          @click="toggle"
          :aria-label="$t('sidebar.newFolder')"
          :title="$t('sidebar.newFolder')"
        >
          <i class="material-icons">{{
            open ? "keyboard_arrow_down" : "keyboard_arrow_up"
          }}</i>
        </button>
      </div>

      <div class="card-content">
        <div class="file" v-for="file in filesInUpload" :key="file.id">
          <div class="file-name">{{ file.name }}</div>
          <div class="file-progress">
            <div v-bind:style="{ width: file.progress + '%' }"></div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { mapGetters } from "vuex";

export default {
  name: "uploadFiles",
  data: function () {
    return {
      open: false,
    };
  },
  computed: {
    ...mapGetters(["filesInUpload", "filesInUploadCount"]),
  },
  methods: {
    toggle: function () {
      this.open = !this.open;
    },
  },
};
</script>
