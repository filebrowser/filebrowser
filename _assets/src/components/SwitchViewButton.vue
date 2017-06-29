<template>
  <button @click="change" aria-label="Switch View" title="Switch View" class="action">
    <i class="material-icons">{{ icon() }}</i>
    <span>Switch view</span>
  </button>
</template>

<script>
import page from '../page'

var $ = window.info

export default {
  name: 'switch-button',
  methods: {
    change: function (event) {
      let url = window.location.pathname + '?display='

      if ($.req.data.display === 'mosaic') {
        url += 'list'
      } else {
        url += 'mosaic'
      }

      page.open(url)
    },
    icon: function () {
      if ($.req.data.display === 'mosaic') return 'view_list'
      return 'view_module'
    }
  }
}
</script>
