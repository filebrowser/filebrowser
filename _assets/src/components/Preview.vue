<template>
  <div id="previewer">
    <div class="bar">
      <button @click="back" class="action" aria-label="Close Preview" id="close">
        <i class="material-icons">close</i>
      </button>

      <rename-button v-if="allowEdit()"></rename-button>
      <delete-button v-if="allowEdit()"></delete-button>
      <download-button></download-button>
      <info-button></info-button>
    </div>

    <div class="preview">
      <img v-if="req.type == 'image'" :src="raw()">
      <audio v-else-if="req.type == 'audio'" :src="raw()" controls></audio>
      <video v-else-if="req.type == 'video'" :src="raw()" controls>
        Sorry, your browser doesn't support embedded videos,
        but don't worry, you can <a :href="download()">download it</a>
        and watch it with your favorite video player!
      </video>
      <object v-else-if="req.extension == '.pdf'" class="pdf" :data="raw()"></object>
      <a v-else-if="req.type == 'blob'" :href="download()">
        <h2 class="message">Download <i class="material-icons">file_download</i></h2>
      </a>
      <pre v-else >{{ req.content }}</pre>
    </div>
  </div>
</template>

<script>
import { mapState } from 'vuex'
import url from '@/utils/url'
import InfoButton from './buttons/InfoButton'
import DeleteButton from './buttons/DeleteButton'
import RenameButton from './buttons/RenameButton'
import DownloadButton from './buttons/DownloadButton'

export default {
  name: 'preview',
  components: {
    InfoButton,
    DeleteButton,
    RenameButton,
    DownloadButton
  },
  computed: mapState(['req']),
  methods: {
    download: function () {
      let url = `${this.$store.state.baseURL}/api/download/`
      url += this.req.url.slice(6)
      url += `?token=${this.$store.state.jwt}`

      return url
    },
    raw: function () {
      return `${this.download()}&inline=true`
    },
    back: function (event) {
      let uri = url.removeLastDir(this.$route.path) + '/'
      this.$router.push({ path: uri })
    },
    allowEdit: function (event) {
      return this.$store.state.user.allowEdit
    }
  }
}
</script>
