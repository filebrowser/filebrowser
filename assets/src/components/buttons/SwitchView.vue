<template>
  <button @click="change" :aria-label="$t('buttons.switchView')" :title="$t('buttons.switchView')" class="action" id="switch-view-button">
    <i class="material-icons">{{ icon() }}</i>
    <span>{{ $t('buttons.switchView') }}</span>
  </button>
</template>

<script>
export default {
  name: 'switch-button',
  methods: {
    change: function (event) {
      // If we are on mobile we should close the dropdown.
      this.$store.commit('closeHovers')

      let display = 'mosaic'

      if (this.$store.state.req.display === 'mosaic') {
        display = 'list'
      }

      this.$store.commit('listingDisplay', display)
      let path = this.$store.state.baseURL
      if (path === '') path = '/'
      document.cookie = `display=${display}; max-age=31536000; path=${path}`
    },
    icon: function () {
      if (this.$store.state.req.display === 'mosaic') return 'view_list'
      return 'view_module'
    }
  }
}
</script>
