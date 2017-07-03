<template>
  <button title="Info" aria-label="Info" class="action" @click="$store.commit('showInfo', true)">
    <i class="material-icons">info</i>
    <span>Info</span>
  </button>
</template>

<script>
export default {name: 'info-button'}
</script>
