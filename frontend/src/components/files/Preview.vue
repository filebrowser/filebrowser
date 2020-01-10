<template>
  <div id="previewer">
    <div class="bar">
      <button @click="back" class="action" :title="$t('files.closePreview')" :aria-label="$t('files.closePreview')" id="close">
        <i class="material-icons">close</i>
      </button>
      <span class="title">{{ req.name }}</span>
      <edit-button v-if="isFileEditable"></edit-button>
      <rename-button v-if="user.perm.rename"></rename-button>
      <delete-button v-if="user.perm.delete"></delete-button>
      <download-button v-if="user.perm.download"></download-button>
      <info-button></info-button>
    </div>

    <button class="action" @click="prev" v-show="hasPrevious" :aria-label="$t('buttons.previous')" :title="$t('buttons.previous')">
      <i class="material-icons">chevron_left</i>
    </button>
    <button class="action" @click="next" v-show="hasNext" :aria-label="$t('buttons.next')" :title="$t('buttons.next')">
      <i class="material-icons">chevron_right</i>
    </button>

    <div class="preview">
      <ExtendedImage v-if="req.type == 'image'" :src="raw"></ExtendedImage>
      <audio v-else-if="req.type == 'audio'" :src="raw" autoplay controls></audio>
      <video v-else-if="req.type == 'video'" :src="raw" autoplay controls>
        <track
          kind="captions"
          v-for="(sub, index) in subtitles"
          :key="index"
          :src="sub"
          :label="'Subtitle ' + index" :default="index === 0">
        Sorry, your browser doesn't support embedded videos,
        but don't worry, you can <a :href="download">download it</a>
        and watch it with your favorite video player!
      </video>
      <object v-else-if="req.extension == '.pdf'" class="pdf" :data="raw"></object>
      <div class="html" v-else-if="req.extension == '.html'"> <iframe :src="getHtmlContent()"></iframe> </div>
      <div class="markdown" v-else-if="isMarkdown" v-html="getMarkdownContent()"></div>
      <div class="text" v-else-if="req.type == 'text'">{{ previewContent }}</div>
      <div v-else>
        <a :href="download">
          <h2 class="message">{{ $t('buttons.download') }} <i class="material-icons">file_download</i></h2>
        </a>
      </div>
    </div>
  </div>
</template>

<script>
import { mapGetters, mapState } from 'vuex'
import url from '@/utils/url'
import { baseURL } from '@/utils/constants'
import { files as api } from '@/api'
import marked from 'marked'
import InfoButton from '@/components/buttons/Info'
import DeleteButton from '@/components/buttons/Delete'
import RenameButton from '@/components/buttons/Rename'
import DownloadButton from '@/components/buttons/Download'
import EditButton from '@/components/buttons/Edit'
import ExtendedImage from './ExtendedImage'

const markdownExtesions = [
  '.md',
  '.mkdn',
  '.mdwn',
  '.mdown',
  '.markdown'
]

export default {
  name: 'preview',
  components: {
    InfoButton,
    DeleteButton,
    RenameButton,
    DownloadButton,
    EditButton,
    ExtendedImage
  },
  data: function () {
    return {
      previousLink: '',
      nextLink: '',
      listing: null,
      subtitles: []
    }
  },
  computed: {
    ...mapGetters(['isFileEditable']),
    ...mapState(['req', 'user', 'oldReq', 'jwt', 'previewContent']),
    hasPrevious () {
      return (this.previousLink !== '')
    },
    hasNext () {
      return (this.nextLink !== '')
    },
    download () {
      return `${baseURL}/api/raw${this.req.path}?auth=${this.jwt}`
    },
    raw () {
      return `${this.download}&inline=true`
    },
    isMarkdown () {
      return markdownExtesions.includes(this.req.extension)
    }
  },
  async mounted () {
    window.addEventListener('keydown', this.key)

    if (this.req.subtitles) {
      this.subtitles = this.req.subtitles.map(sub => `${baseURL}/api/raw${sub}?auth=${this.jwt}&inline=true`)
    }

    try {
      if (this.oldReq.items) {
        this.updateLinks(this.oldReq.items)
      } else {
        const path = url.removeLastDir(this.$route.path)
        const res = await api.fetch(path)
        this.updateLinks(res.items)
      }
    } catch (e) {
      this.$showError(e)
    }
  },
  beforeDestroy () {
    window.removeEventListener('keydown', this.key)
  },
  methods: {
    back () {
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
      } else if (event.which == 27) { // esc
        this.back()
      }
    },
    updateLinks (items) {
      for (let i = 0; i < items.length; i++) {
        if (items[i].name !== this.req.name) {
          continue
        }

        for (let j = i - 1; j >= 0; j--) {
          if (!items[j].isDir) {
            this.previousLink = items[j].url
            break
          }
        }

        for (let j = i + 1; j < items.length; j++) {
          if (!items[j].isDir) {
            this.nextLink = items[j].url
            break
          }
        }

        return
      }
    },
    getHtmlContent () {
      return 'data:text/html;base64,' + btoa(unescape(encodeURIComponent(this.previewContent)))
    },
    getMarkdownContent () {
      return marked(this.previewContent)
    }
  }
}
</script>
