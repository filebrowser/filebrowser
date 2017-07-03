<template>
  <button @click="download" aria-label="Download" title="Download" class="action">
    <i class="material-icons">file_download</i>
    <span>Download</span>
    <span v-if="selectedCount > 0" class="counter">{{ selectedCount }}</span>
  </button>
</template>

<script>
import {mapGetters, mapState} from 'vuex'

export default {
  name: 'download-button',
  computed: {
    ...mapState(['req']),
    ...mapGetters(['selectedCount'])
  },
  methods: {
    download: function (event) {
      if (this.req.kind !== 'listing') {
        let url = this.$route.params[0]
        url = this.$store.state.baseURL + '/api/download/' + url
        url += '?token=' + this.$store.state.jwt
        window.open(url)
        return
      }

      this.$store.commit('showPrompt', 'download')
    }
  }
}
</script>
