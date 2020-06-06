<template>
  <button id="move-button" :aria-label="$t('buttons.move')" :title="$t('buttons.move')" class="action" @click="show">
    <i class="material-icons">forward</i>
    <span>{{ $t('buttons.moveFile') }}</span>
  </button>
</template>

<script>
export default {
  name: 'MoveButton',
  methods: {
    show: function() {
      this.$store.commit('showHover', 'move')
    }
  }
}
</script>
