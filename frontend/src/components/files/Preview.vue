<template>
  <div id="previewer" @mousemove="toggleNavigation" @touchstart="toggleNavigation">
    <header-bar>
      <action icon="close" :label="$t('buttons.close')" @action="close()" />
      <title>{{ name }}</title>
      <preview-size-button v-if="isResizeEnabled && req.type === 'image'" @change-size="toggleSize" v-bind:size="fullSize" :disabled="loading" />

      <template #actions>
        <rename-button :disabled="loading" v-if="user.perm.rename" />
        <delete-button :disabled="loading" v-if="user.perm.delete" />
        <download-button :disabled="loading" v-if="user.perm.download" />
        <info-button :disabled="loading" />
      </template>
    </header-bar>

    <div class="loading" v-if="loading">
      <div class="spinner">
        <div class="bounce1"></div>
        <div class="bounce2"></div>
        <div class="bounce3"></div>
      </div>
    </div>

    <template v-if="!loading">
      <div class="preview">
        <ExtendedImage v-if="req.type == 'image'" :src="raw"></ExtendedImage>
        <audio v-else-if="req.type == 'audio'" :src="raw" controls></audio>
        <video v-else-if="req.type == 'video'" :src="raw" controls>
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
        <object v-else-if="req.extension.toLowerCase() == '.pdf'" class="pdf" :data="raw"></object>
        <a v-else-if="req.type == 'blob'" :href="download">
          <h2 class="message">{{ $t('buttons.download') }} <i class="material-icons">file_download</i></h2>
        </a>
      </div>
    </template>

    <button @click="prev" @mouseover="hoverNav = true" @mouseleave="hoverNav = false" :class="{ hidden: !hasPrevious || !showNav }" :aria-label="$t('buttons.previous')" :title="$t('buttons.previous')">
      <i class="material-icons">chevron_left</i>
    </button>
    <button @click="next" @mouseover="hoverNav = true" @mouseleave="hoverNav = false" :class="{ hidden: !hasNext || !showNav }" :aria-label="$t('buttons.next')" :title="$t('buttons.next')">
      <i class="material-icons">chevron_right</i>
    </button>
  </div>
</template>

<script>
import { mapState } from 'vuex'
import { files as api } from '@/api'
import { baseURL, resizePreview } from '@/utils/constants'
import url from '@/utils/url'
import throttle from 'lodash.throttle'

import HeaderBar from '@/components/header/HeaderBar'
import Action from '@/components/header/Action'
import PreviewSizeButton from '@/components/buttons/PreviewSize'
import InfoButton from '@/components/buttons/Info'
import DeleteButton from '@/components/buttons/Delete'
import RenameButton from '@/components/buttons/Rename'
import DownloadButton from '@/components/buttons/Download'
import ExtendedImage from './ExtendedImage'

const mediaTypes = [
  "image",
  "video",
  "audio",
  "blob"
]

export default {
  name: 'preview',
  components: {
    HeaderBar,
    Action,
    PreviewSizeButton,
    InfoButton,
    DeleteButton,
    RenameButton,
    DownloadButton,
    ExtendedImage
  },
  data: function () {
    return {
      previousLink: '',
      nextLink: '',
      listing: null,
      name: '',
      subtitles: [],
      fullSize: false,
      showNav: true,
      navTimeout: null,
      hoverNav: false
    }
  },
  computed: {
    ...mapState(['req', 'user', 'oldReq', 'jwt', 'loading', 'show']),
    hasPrevious () {
      return (this.previousLink !== '')
    },
    hasNext () {
      return (this.nextLink !== '')
    },
    download () {
      return `${baseURL}/api/raw${url.encodePath(this.req.path)}?auth=${this.jwt}`
    },
    previewUrl () {
      if (this.req.type === 'image' && !this.fullSize) {
        return `${baseURL}/api/preview/big${url.encodePath(this.req.path)}?auth=${this.jwt}`
      }
      return `${baseURL}/api/raw${url.encodePath(this.req.path)}?auth=${this.jwt}`
    },
    raw () {
      return `${this.previewUrl}&inline=true`
    },
    showMore () {
      return this.$store.state.show === 'more'
    },
    isResizeEnabled () {
      return resizePreview
    }
  },
  watch: {
    $route: function () {
      this.updatePreview()
      this.toggleNavigation()
    }
  },
  async mounted () {
    window.addEventListener('keydown', this.key)
    this.$store.commit('setPreviewMode', true)
    this.listing = this.oldReq.items
    this.$root.$on('preview-deleted', this.deleted)
    this.updatePreview()
  },
  beforeDestroy () {
    window.removeEventListener('keydown', this.key)
    this.$store.commit('setPreviewMode', false)
    this.$root.$off('preview-deleted', this.deleted)
  },
  methods: {
    deleted () {
      this.listing = this.listing.filter(item => item.name !== this.name)

      if (this.hasNext) {
        this.next()
      } else if (!this.hasPrevious && !this.hasNext) {
        this.back()
      } else {
        this.prev()
      }
    },
    back () {
      this.$store.commit('setPreviewMode', false)
      let uri = url.removeLastDir(this.$route.path) + '/'
      this.$router.push({ path: uri })
    },
    prev () {
      this.hoverNav = false
      this.$router.push({ path: this.previousLink })
    },
    next () {
      this.hoverNav = false
      this.$router.push({ path: this.nextLink })
    },
    key (event) {

      if (this.show !== null) {
        return
      }

      if (event.which === 13 || event.which === 39) { // right arrow
        if (this.hasNext) this.next()
      } else if (event.which === 37) { // left arrow
        if (this.hasPrevious) this.prev()
      } else if (event.which === 27) { // esc
        this.back()
      }
    },
    async updatePreview () {
      if (this.req.subtitles) {
        this.subtitles = this.req.subtitles.map(sub => `${baseURL}/api/raw${sub}?auth=${this.jwt}&inline=true`)
      }

      let dirs = this.$route.fullPath.split("/")
      this.name = decodeURIComponent(dirs[dirs.length - 1])

      if (!this.listing) {
        try {
          const path = url.removeLastDir(this.$route.path)
          const res = await api.fetch(path)
          this.listing = res.items
        } catch (e) {
          this.$showError(e)
        }
      }

      this.previousLink = ''
      this.nextLink = ''

      for (let i = 0; i < this.listing.length; i++) {
        if (this.listing[i].name !== this.name) {
          continue
        }

        for (let j = i - 1; j >= 0; j--) {
          if (mediaTypes.includes(this.listing[j].type)) {
            this.previousLink = this.listing[j].url
            break
          }
        }

        for (let j = i + 1; j < this.listing.length; j++) {
          if (mediaTypes.includes(this.listing[j].type)) {
            this.nextLink = this.listing[j].url
            break
          }
        }

        return
      }
    },
    openMore () {
      this.$store.commit('showHover', 'more')
    },
    resetPrompts () {
      this.$store.commit('closeHovers')
    },
    toggleSize () {
      this.fullSize = !this.fullSize
    },
    toggleNavigation: throttle(function() {
      this.showNav = true

      if (this.navTimeout) {
        clearTimeout(this.navTimeout)
      }

      this.navTimeout = setTimeout(() => {
        this.showNav = false || this.hoverNav
        this.navTimeout = null
      }, 1500);
    }, 500),
    close () {
      let uri = url.removeLastDir(this.$route.path) + '/'
      this.$router.push({ path: uri })
    }
  }
}
</script>
