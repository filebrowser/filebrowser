<template>
  <button @click="$store.commit('showRename', true)" aria-label="Rename" title="Rename" class="action">
    <i class="material-icons">mode_edit</i>
    <span>Rename</span>
  </button>
</template>

<script>
export default {name: 'rename-button'}
</script>
