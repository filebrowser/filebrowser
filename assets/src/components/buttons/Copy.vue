<template>
  <button @click="show" aria-label="Copy" title="Copy" class="action" id="copy-button">
    <i class="material-icons">content_copy</i>
    <span>Copy file</span>
  </button>
</template>

<script>
export default {
  name: 'copy-button',
  methods: {
    show: function (event) {
      this.$store.commit('showHover', 'copy')
    }
  }
}
</script>
