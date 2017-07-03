<template>
  <button @click="this.$store.commit('showPrompt', 'move')" aria-label="Move" title="Move" class="action">
    <i class="material-icons">forward</i>
    <span>Move file</span>
  </button>
</template>

<script>
export default {name: 'move-button'}
</script>
