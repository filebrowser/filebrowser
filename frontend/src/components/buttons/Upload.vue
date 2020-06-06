<template>
  <button id="upload-button" :aria-label="$t('buttons.upload')" :title="$t('buttons.upload')" class="action" @click="upload">
    <i class="material-icons">file_upload</i>
    <span>{{ $t('buttons.upload') }}</span>
  </button>
</template>

<script>
export default {
  name: 'UploadButton',
  methods: {
    upload: function() {
      document.getElementById('upload-input').click()
    }
  }
}
</script>
