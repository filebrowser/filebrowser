<template>
  <button @click="download" :aria-label="$t('buttons.download')" :title="$t('buttons.download')" id="download-button" class="action">
    <i class="material-icons">file_download</i>
    <span>{{ $t('buttons.download') }}</span>
    <span v-if="selectedCount > 0" class="counter">{{ selectedCount }}</span>
  </button>
</template>

<script>
import {mapGetters, mapState} from 'vuex'
import { files as api } from '@/api'

export default {
  name: 'download-button',
  computed: {
    ...mapState(['req', 'selected']),
    ...mapGetters(['isListing', 'selectedCount'])
  },
  methods: {
    download: function () {
      if (!this.isListing) {
        api.download(null, this.$route.path)
        return
      }

      if (this.selectedCount === 1 && !this.req.items[this.selected[0]].isDir) {
        api.download(null, this.req.items[this.selected[0]].url)
        return
      }

      this.$store.commit('showHover', 'download')
    }
  }
}
</script>
