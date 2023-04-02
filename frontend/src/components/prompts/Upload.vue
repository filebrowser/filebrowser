<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t("prompts.upload") }}</h2>
    </div>

    <div class="card-content">
      <p>{{ $t("prompts.uploadMessage") }}</p>
    </div>

    <div class="card-action full">
      <div @click="uploadFile" class="action">
        <i class="material-icons">insert_drive_file</i>
        <div class="title">{{ $t("buttons.file") }}</div>
      </div>
      <div @click="uploadFolder" class="action">
        <i class="material-icons">folder</i>
        <div class="title">{{ $t("buttons.folder") }}</div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: "upload",
  methods: {
    uploadFile: function () {
      document.getElementById("upload-input").value = "";
      document.getElementById("upload-input").click();
    },
    uploadFolder: function () {
      document.getElementById("upload-folder-input").value = "";
      document.getElementById("upload-folder-input").click();
    },
  },
};
</script>
