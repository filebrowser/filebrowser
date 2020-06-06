<template>
  <button :aria-label="$t('buttons.share')" :title="$t('buttons.share')" class="action" @click="show">
    <i class="material-icons">share</i>
    <span>{{ $t('buttons.share') }}</span>
  </button>
</template>

<script>
export default {
  name: 'ShareButton',
  methods: {
    show() {
      this.$store.commit('showHover', 'share')
    }
  }
}
</script>
