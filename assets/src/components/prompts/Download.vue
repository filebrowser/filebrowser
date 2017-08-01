<template>
  <div class="prompt" id="download">
    <h3>{{ $t('prompts.download') }}</h3>
    <p>{{ $t('prompts.downloadMessage') }}</p>

    <button @click="download('zip')" autofocus>zip</button>
    <button @click="download('tar')" autofocus>tar</button>
    <button @click="download('targz')" autofocus>tar.gz</button>
    <button @click="download('tarbz2')" autofocus>tar.bz2</button>
    <button @click="download('tarxz')" autofocus>tar.xz</button>
  </div>
</template>

<script>
import {mapGetters, mapState} from 'vuex'
import api from '@/utils/api'

export default {
  name: 'download',
  computed: {
    ...mapState(['selected', 'req']),
    ...mapGetters(['selectedCount'])
  },
  methods: {
    download: function (format) {
      if (this.selectedCount === 0) {
        api.download(format, this.$route.path)
      } else {
        let files = []

        for (let i of this.selected) {
          files.push(this.req.items[i].url)
        }

        api.download(format, ...files)
      }

      this.$store.commit('closeHovers')
    }
  }
}
</script>
