<template>
  <div id="previewer">
    <div class="bar">
      <button @click="back" class="action" :title="$t('files.closePreview')" :aria-label="$t('files.closePreview')" id="close">
        <i class="material-icons">close</i>
      </button>

      <rename-button v-if="allowEdit()"></rename-button>
      <delete-button v-if="allowEdit()"></delete-button>
      <download-button></download-button>
      <info-button></info-button>
    </div>

    <button class="action" @click="prev" v-show="hasPrevious" :aria-label="$t('buttons.previous')" :title="$t('buttons.previous')">
      <i class="material-icons">chevron_left</i>
    </button>
    <button class="action" @click="next" v-show="hasNext" :aria-label="$t('buttons.next')" :title="$t('buttons.next')">
      <i class="material-icons">chevron_right</i>
    </button>

    <div class="preview">
      <img v-if="req.type == 'image'" :src="raw()">
      <audio v-else-if="req.type == 'audio'" :src="raw()" autoplay controls></audio>
      <video v-else-if="req.type == 'video'" :src="raw()" autoplay controls>
        Sorry, your browser doesn't support embedded videos,
        but don't worry, you can <a :href="download()">download it</a>
        and watch it with your favorite video player!
      </video>
      <object v-else-if="req.extension == '.pdf'" class="pdf" :data="raw()"></object>
      <a v-else-if="req.type == 'blob'" :href="download()">
        <h2 class="message">{{ $t('buttons.download') }} <i class="material-icons">file_download</i></h2>
      </a>
      <pre v-else >{{ req.content }}</pre>
    </div>
  </div>
</template>

<script>
import { mapState } from 'vuex'
import url from '@/utils/url'
import * as api from '@/utils/api'
import InfoButton from '@/components/buttons/Info'
import DeleteButton from '@/components/buttons/Delete'
import RenameButton from '@/components/buttons/Rename'
import DownloadButton from '@/components/buttons/Download'

export default {
  name: 'preview',
  components: {
    InfoButton,
    DeleteButton,
    RenameButton,
    DownloadButton
  },
  data: function () {
    return {
      previousLink: '',
      nextLink: '',
      listing: null
    }
  },
  computed: {
    ...mapState(['req', 'oldReq']),
    hasPrevious () {
      return (this.previousLink !== '')
    },
    hasNext () {
      return (this.nextLink !== '')
    }
  },
  mounted () {
    window.addEventListener('keyup', this.key)
    api.fetch(url.removeLastDir(this.$route.path))
      .then(req => {
        this.listing = req
        this.updateLinks()
      })
      .catch(this.$showError)
  },
  beforeDestroy () {
    window.removeEventListener('keyup', this.key)
  },
  methods: {
    download () {
      let url = `${this.$store.state.baseURL}/api/download`
      url += this.req.url.slice(6)

      return url
    },
    raw () {
      return `${this.download()}?&inline=true`
    },
    back (event) {
      let uri = url.removeLastDir(this.$route.path) + '/'
      this.$router.push({ path: uri })
    },
    prev () {
      this.$router.push({ path: this.previousLink })
    },
    next () {
      this.$router.push({ path: this.nextLink })
    },
    key (event) {
      event.preventDefault()

      if (event.which === 13 || event.which === 39) { // right arrow
        if (this.hasNext) this.next()
      } else if (event.which === 37) { // left arrow
        if (this.hasPrevious) this.prev()
      }
    },
    updateLinks () {
      let pos = null

      for (let i = 0; i < this.listing.items.length; i++) {
        if (this.listing.items[i].name === this.req.name) {
          pos = i
          break
        }
      }

      if (pos === null) {
        return
      }

      if (pos !== 0) {
        this.previousLink = this.listing.items[pos - 1].url
      }

      if (pos !== this.listing.items.length - 1) {
        this.nextLink = this.listing.items[pos + 1].url
      }
    },
    allowEdit (event) {
      return this.$store.state.user.allowEdit
    }
  }
}
</script>
