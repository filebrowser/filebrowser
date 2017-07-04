<template>
  <button @click="show" aria-label="Move" title="Move" class="action">
    <i class="material-icons">forward</i>
    <span>Move file</span>
  </button>
</template>

<script>
export default {
  name: 'move-button',
  methods: {
    show: function (event) {
      this.$store.commit('showHover', 'move')
    }
  }
}
</script>
