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
            <img v-if="type == 'image'" :src="raw()">
            <audio v-else-if="type == 'audio'" :src="raw()" controls></audio>
            <video v-else-if="type == 'video'" :src="raw()" controls>
                Sorry, your browser doesn't support embedded videos,
                but don't worry, you can <a href="?download=true">download it</a>
                and watch it with your favorite video player!
            </video>
            <object v-else-if="extension == '.pdf'" class="pdf" :data="raw()"></object>
            <a v-else-if="type == 'blob'" href="?download=true"><h2 class="message">Download <i class="material-icons">file_download</i></h2></a>
            <pre v-else >{{ content }}</pre>
        </div>
    </div>
</template>

<script>
import page from '../page'
import InfoButton from './InfoButton'
import DeleteButton from './DeleteButton'
import RenameButton from './RenameButton'
import DownloadButton from './DownloadButton'

export default {
  name: 'preview',
  components: {
    InfoButton,
    DeleteButton,
    RenameButton,
    DownloadButton
  },
  data: function () {
    return window.info.req.data
  },
  methods: {
    raw: function () {
      return this.url + '?raw=true'
    },
    back: function (event) {
      let url = page.removeLastDir(window.location.pathname)
      page.open(url)
    },
    allowEdit: function (event) {
      return window.info.user.allowEdit
    }
  }
}
</script>
