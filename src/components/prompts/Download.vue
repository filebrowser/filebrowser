<template>
  <div class="card floating" id="download">
    <div class="card-title">
      <h2>{{ $t('prompts.download') }}</h2>
    </div>

    <div class="card-content">
      <p>{{ $t('prompts.downloadMessage') }}</p>

      <button class="block cancel" @click="download('zip')" v-focus>zip</button>
      <button class="block cancel" @click="download('tar')" v-focus>tar</button>
      <button class="block cancel" @click="download('targz')" v-focus>tar.gz</button>
      <button class="block cancel" @click="download('tarbz2')" v-focus>tar.bz2</button>
      <button class="block cancel" @click="download('tarxz')" v-focus>tar.xz</button>
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
