<template>
    <button @click="show" aria-label="Rename" title="Rename" class="action">
        <i class="material-icons">mode_edit</i>
        <span>Rename</span>
    </button>
</template>

<script>
export default {
  name: 'rename-button',
  methods: {
    show: function (event) {
      this.$store.commit('showRename', true)
    }
  }
}
</script>
