<template>
  <button @click="show" :aria-label="$t('buttons.share')" :title="$t('buttons.share')" class="action">
    <i class="material-icons">share</i>
    <span>{{ $t('buttons.share') }}</span>
  </button>
</template>

<script>
export default {
  name: 'share-button',
  methods: {
    show () {
      this.$store.commit('showHover', 'share')
    }
  }
}
</script>
