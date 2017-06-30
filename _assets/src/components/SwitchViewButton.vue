<template>
  <button @click="change" aria-label="Switch View" title="Switch View" class="action">
    <i class="material-icons">{{ icon() }}</i>
    <span>Switch view</span>
  </button>
</template>

<script>
import page from '../utils/page'

export default {
  name: 'switch-button',
  methods: {
    change: function (event) {
      let url = window.location.pathname + '?display='

      if (this.$store.state.req.data.display === 'mosaic') {
        url += 'list'
      } else {
        url += 'mosaic'
      }

      page.open(url)
    },
    icon: function () {
      if (this.$store.state.req.data.display === 'mosaic') return 'view_list'
      return 'view_module'
    }
  }
}
</script>
