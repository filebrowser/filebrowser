<template>
  <button @click="upload" aria-label="Upload" title="Upload" class="action">
    <i class="material-icons">file_upload</i>
    <span>Upload</span>
  </button>
</template>

<script>
export default {
  name: 'upload-button',
  methods: {
    upload: function (event) {
      document.getElementById('upload-input').click()
    }
  }
}
</script>
