<template>
  <button @click="download" aria-label="Download" title="Download" id="download-button" class="action">
    <i class="material-icons">file_download</i>
    <span>Download</span>
    <span v-if="selectedCount > 0" class="counter">{{ selectedCount }}</span>
  </button>
</template>

<script>
import {mapGetters, mapState} from 'vuex'
import api from '@/utils/api'

export default {
  name: 'download-button',
  computed: {
    ...mapState(['req', 'selected']),
    ...mapGetters(['selectedCount'])
  },
  methods: {
    download: function (event) {
      if (this.req.kind !== 'listing') {
        api.download(null, this.$route.path)
        return
      }

      if (this.selectedCount === 1) {
        api.download(null, this.req.items[this.selected[0]].url)
        return
      }

      this.$store.commit('showHover', 'download')
    }
  }
}
</script>
