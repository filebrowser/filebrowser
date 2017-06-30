<template>
  <div class="prompt" id="download">
    <h3>Download files</h3>
    <p>Choose the format you want to download.</p>
    <button @click="download('zip')" autofocus>zip</button>
    <button @click="download('tar')" autofocus>tar</button>
    <button @click="download('targz')" autofocus>tar.gz</button>
    <button @click="download('tarbz2')" autofocus>tar.bz2</button>
    <button @click="download('tarxz')" autofocus>tar.xz</button>
  </div>
</template>

<script>
import {mapGetters, mapState} from 'vuex'

export default {
  name: 'download-prompt',
  computed: {
    ...mapState(['selected', 'req']),
    ...mapGetters(['selectedCount'])
  },
  methods: {
    download: function (format) {
      let uri = `${window.location.pathname}?download=${format}`

      if (this.selectedCount > 0) {
        let files = ''

        for (let i of this.selected) {
          files += this.req.data.items[i].url.replace(window.location.pathname, '') + ','
        }

        files = files.substring(0, files.length - 1)
        files = encodeURIComponent(files)
        uri += `&files=${files}`
      }

      window.open(uri)
    }
  }
}
</script>
