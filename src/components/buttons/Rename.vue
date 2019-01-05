<template>
  <button @click="show" :aria-label="$t('buttons.rename')" :title="$t('buttons.rename')" class="action" id="rename-button">
    <i class="material-icons">mode_edit</i>
    <span>{{ $t('buttons.rename') }}</span>
  </button>
</template>

<script>
export default {
  name: 'rename-button',
  methods: {
    show: function () {
      this.$store.commit('showHover', 'rename')
    }
  }
}
</script>
