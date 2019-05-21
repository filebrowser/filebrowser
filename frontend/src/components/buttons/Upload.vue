<template>
  <button @click="upload" :aria-label="$t('buttons.upload')" :title="$t('buttons.upload')" class="action" id="upload-button">
    <i class="material-icons">file_upload</i>
    <span>{{ $t('buttons.upload') }}</span>
  </button>
</template>

<script>
export default {
  name: 'upload-button',
  methods: {
    upload: function () {
      document.getElementById('upload-input').click()
    }
  }
}
</script>
