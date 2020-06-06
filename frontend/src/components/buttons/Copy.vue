<template>
  <button id="copy-button" :aria-label="$t('buttons.copy')" :title="$t('buttons.copy')" class="action" @click="show">
    <i class="material-icons">content_copy</i>
    <span>{{ $t('buttons.copyFile') }}</span>
  </button>
</template>

<script>
export default {
  name: 'CopyButton',
  methods: {
    show: function() {
      this.$store.commit('showHover', 'copy')
    }
  }
}
</script>
