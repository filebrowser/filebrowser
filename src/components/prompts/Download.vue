<template>
  <div class="card floating" id="download">
    <div class="card-title">
      <h2>{{ $t('prompts.download') }}</h2>
    </div>

    <div class="card-content">
      <p>{{ $t('prompts.downloadMessage') }}</p>

      <button class="block cancel" @click="download('zip')" autofocus>zip</button>
      <button class="block cancel" @click="download('tar')" autofocus>tar</button>
      <button class="block cancel" @click="download('targz')" autofocus>tar.gz</button>
      <button class="block cancel" @click="download('tarbz2')" autofocus>tar.bz2</button>
      <button class="block cancel" @click="download('tarxz')" autofocus>tar.xz</button>
    </div>
  </div>
</template>

<script>
import {mapGetters, mapState} from 'vuex'
import * as api from '@/utils/api'

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
