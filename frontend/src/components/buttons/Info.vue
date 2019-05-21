<template>
  <button :title="$t('buttons.info')" :aria-label="$t('buttons.info')" class="action" @click="show">
    <i class="material-icons">info</i>
    <span>{{ $t('buttons.info') }}</span>
  </button>
</template>

<script>
export default {
  name: 'info-button',
  methods: {
    show: function () {
      this.$store.commit('showHover', 'info')
    }
  }
}
</script>
