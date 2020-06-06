<template>
  <button id="rename-button" :aria-label="$t('buttons.rename')" :title="$t('buttons.rename')" class="action" @click="show">
    <i class="material-icons">mode_edit</i>
    <span>{{ $t('buttons.rename') }}</span>
  </button>
</template>

<script>
export default {
  name: 'RenameButton',
  methods: {
    show: function() {
      this.$store.commit('showHover', 'rename')
    }
  }
}
</script>
