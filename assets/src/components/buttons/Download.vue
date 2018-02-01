<template>
  <button @click="download" :aria-label="$t('buttons.download')" :title="$t('buttons.download')" id="download-button" class="action">
    <i class="material-icons">file_download</i>
    <span>{{ $t('buttons.download') }}</span>
    <span v-if="selectedCount > 0" class="counter">{{ selectedCount }}</span>
  </button>
</template>

<script>
import {mapGetters, mapState} from 'vuex'
import * as api from '@/utils/api'

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
      if (this.selectedCount === 1 && !this.req.items[this.selected[0]].isDir) {
        api.download(null, this.req.items[this.selected[0]].url)
        return
      }

      // Otherwise show the prompt to choose the formt of the download.
      this.$store.commit('showHover', 'download')
    }
  }
}
</script>
