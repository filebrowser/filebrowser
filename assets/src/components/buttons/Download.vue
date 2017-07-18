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
      // If we are not on a listing, download the current file.
      if (this.req.kind !== 'listing') {
        api.download(null, this.$route.path)
        return
      }

      // If we are on a listing and there is one element selected,
      // download it.
      if (this.selectedCount === 1) {
        api.download(null, this.req.items[this.selected[0]].url)
        return
      }

      // Otherwise show the prompt to choose the formt of the download.
      this.$store.commit('showHover', 'download')
    }
  }
}
</script>
