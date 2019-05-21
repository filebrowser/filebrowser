<template>
  <button @click="show" :aria-label="$t('buttons.copy')" :title="$t('buttons.copy')" class="action" id="copy-button">
    <i class="material-icons">content_copy</i>
    <span>{{ $t('buttons.copyFile') }}</span>
  </button>
</template>

<script>
export default {
  name: 'copy-button',
  methods: {
    show: function () {
      this.$store.commit('showHover', 'copy')
    }
  }
}
</script>
