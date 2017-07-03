<template>
  <button @click="change" aria-label="Switch View" title="Switch View" class="action">
    <i class="material-icons">{{ icon() }}</i>
    <span>Switch view</span>
  </button>
</template>

<script>
export default {
  name: 'switch-button',
  methods: {
    change: function (event) {
      let display = 'mosaic'

      if (this.$store.state.req.display === 'mosaic') {
        display = 'list'
      }

      this.$store.commit('listingDisplay', display)
      document.cookie = `display=${display}; max-age=31536000; path=${this.$store.state.baseURL}`
    },
    icon: function () {
      if (this.$store.state.req.display === 'mosaic') return 'view_list'
      return 'view_module'
    }
  }
}
</script>
