<template>
  <button @click="show" aria-label="Rename" title="Rename" class="action" id="rename-button">
    <i class="material-icons">mode_edit</i>
    <span>Rename</span>
  </button>
</template>

<script>
export default {
  name: 'rename-button',
  methods: {
    show: function (event) {
      this.$store.commit('showHover', 'rename')
    }
  }
}
</script>
