<template>
  <div id="download" class="card floating">
    <div class="card-title">
      <h2>{{ $t('prompts.download') }}</h2>
    </div>

    <div class="card-content">
      <p>{{ $t('prompts.downloadMessage') }}</p>

      <button v-focus class="button button--block" @click="download('zip')">zip</button>
      <button v-focus class="button button--block" @click="download('tar')">tar</button>
      <button v-focus class="button button--block" @click="download('targz')">tar.gz</button>
      <button v-focus class="button button--block" @click="download('tarbz2')">tar.bz2</button>
      <button v-focus class="button button--block" @click="download('tarxz')">tar.xz</button>
      <button v-focus class="button button--block" @click="download('tarlz4')">tar.lz4</button>
      <button v-focus class="button button--block" @click="download('tarsz')">tar.sz</button>
    </div>
  </div>
</template>

<script>
import { mapGetters, mapState } from 'vuex'
import { files as api } from '@/api'

export default {
  name: 'Download',
  computed: {
    ...mapState(['selected', 'req']),
    ...mapGetters(['selectedCount'])
  },
  methods: {
    download: function(format) {
      if (this.selectedCount === 0) {
        api.download(format, this.$route.path)
      } else {
        const files = []

        for (const i of this.selected) {
          files.push(this.req.items[i].url)
        }

        api.download(format, ...files)
      }

      this.$store.commit('closeHovers')
    }
  }
}
</script>
