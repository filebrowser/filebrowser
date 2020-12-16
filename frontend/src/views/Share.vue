<template>
  <div v-if="loaded">
    <div id="breadcrumbs">
      <router-link :to="'/share/' + hash" :aria-label="$t('files.home')" :title="$t('files.home')">
        <i class="material-icons">home</i>
      </router-link>

      <span v-for="(link, index) in breadcrumbs" :key="index">
          <span class="chevron"><i class="material-icons">keyboard_arrow_right</i></span>
          <router-link :to="link.url">{{ link.name }}</router-link>
        </span>
    </div>
    <div class="share">
      <div class="share__box share__box__info">
          <div class="share__box__header">
            {{ file.isDir ? sharedSelectedCount > 0 ? $t('download.downloadSelected') : $t('download.downloadFolder') : $t('download.downloadFile') }}
          </div>
          <div class="share__box__element share__box__center share__box__icon">
            <i class="material-icons">{{ icon }}</i>
          </div>
          <div class="share__box__element">
            <strong>{{ $t('prompts.displayName') }}</strong> {{ file.name }}
          </div>
          <div class="share__box__element">
            <strong>{{ $t('prompts.lastModified') }}:</strong> {{ humanTime }}
          </div>
          <div class="share__box__element">
            <strong>{{ $t('prompts.size') }}:</strong> {{ humanSize }}
          </div>
          <div class="share__box__element share__box__center">
            <a target="_blank" :href="link" class="button button--flat">{{ $t('buttons.download') }}</a>
          </div>
          <div class="share__box__element share__box__center">
            <qrcode-vue :value="fullLink" size="200" level="M"></qrcode-vue>
          </div>
          <div v-if="file.isDir" class="share__box__element share__box__center">
            <label>
              <input type="checkbox" :checked="shared.multiple" @click="toggleMultipleSelection">
              {{ $t('buttons.selectMultiple') }}
            </label>
          </div>
      </div>
      <div v-if="file.isDir" class="share__box share__box__items">
        <div class="share__box__header" v-if="file.isDir">
          {{ $t('files.files') }}
        </div>
        <div id="listing" class="list">
          <shared-item v-for="(item) in file.items.slice(0, this.showLimit)"
            :key="base64(item.name)"
            v-bind:index="item.index"
            v-bind:name="item.name"
            v-bind:isDir="item.isDir"
            v-bind:url="item.url"
            v-bind:modified="item.modified"
            v-bind:type="item.type"
            v-bind:size="item.size">
          </shared-item>
          <div v-if="file.items.length > showLimit" class="item">
            <div>
              <p class="name"> + {{ file.items.length - showLimit }} </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import {mapState, mapMutations, mapGetters} from 'vuex';
import { share as api } from '@/api'
import { baseURL } from '@/utils/constants'
import filesize from 'filesize'
import moment from 'moment'
import QrcodeVue from 'qrcode.vue'
import SharedItem from "@/components/files/SharedItem"

export default {
  name: 'share',
  components: {
    SharedItem,
    QrcodeVue
  },
  data: () => ({
    loaded: false,
    notFound: false,
    file: null,
    showLimit: 500
  }),
  watch: {
    '$route': 'fetchData'
  },
  created: async function () {
    await this.fetchData()
  },
  mounted () {
    window.addEventListener('keydown', this.keyEvent)
  },
  beforeDestroy () {
    window.removeEventListener('keydown', this.keyEvent)
  },
  computed: {
    ...mapState(['shared']),
    ...mapGetters(['sharedSelectedCount']),
    icon: function () {
      if (this.file.isDir) return 'folder'
      if (this.file.type === 'image') return 'insert_photo'
      if (this.file.type === 'audio') return 'volume_up'
      if (this.file.type === 'video') return 'movie'
      return 'insert_drive_file'
    },
    hash: function () {
      return this.$route.params.pathMatch.split('/')[0]
    },
    path: function () {
      let absoluteParts = this.file.path.split('/')
      let urlParts = this.$route.params.pathMatch.split('/')

      absoluteParts.shift()

      absoluteParts.forEach((_, i) => absoluteParts[i] = encodeURIComponent(absoluteParts[i]))
      urlParts.forEach((_, i) => urlParts[i] = encodeURIComponent(urlParts[i]))

      if (absoluteParts[absoluteParts.length - 1] === '') absoluteParts.pop()
      if (urlParts[urlParts.length - 1] === '') urlParts.pop()

      if (urlParts.length === 1) return absoluteParts[absoluteParts.length - 1]

      let len = Math.min(absoluteParts.length, urlParts.length)
      for (let i = 0; i < len; i++) {
        if (urlParts[urlParts.length - 1 - i] !== absoluteParts[absoluteParts.length - 1 - i]) return urlParts.slice(urlParts.length - i).join('/')
      }
      return absoluteParts.slice(absoluteParts.length - len).join('/')
    },
    link: function () {
      if (this.sharedSelectedCount === 0) return `${baseURL}/api/public/dl/${this.hash}/${this.path}`
      if (this.sharedSelectedCount === 1) return `${baseURL}/api/public/dl/${this.hash}/${this.path}/${encodeURIComponent(this.file.items[this.shared.selected[0]].name)}`
      let files = []
      for (let s of this.shared.selected) {
        files.push(encodeURIComponent(this.file.items[s].name))
      }
      return `${baseURL}/api/public/dl/${this.hash}/${this.path}/?files=${encodeURIComponent(files.join(','))}`
    },
    fullLink: function () {
      return window.location.origin + this.link
    },
    humanSize: function () {
      if (this.file.isDir) {
        return this.file.items.length
      }

      return filesize(this.file.size)
    },
    humanTime: function () {
      return moment(this.file.modified).fromNow()
    },
    breadcrumbs () {
      let parts = this.path.split('/')

      if (parts[0] === '') {
        parts.shift()
      }

      if (parts[parts.length - 1] === '') {
        parts.pop()
      }

      let breadcrumbs = []

      for (let i = 0; i < parts.length; i++) {
        if (i === 0) {
          breadcrumbs.push({ name: decodeURIComponent(parts[i]), url: '/share/' + this.hash + '/' + parts[i] + '/' })
        } else  {
          breadcrumbs.push({ name: decodeURIComponent(parts[i]), url: breadcrumbs[i - 1].url + parts[i] + '/' })
        }
      }

      breadcrumbs.shift()

      if (breadcrumbs.length > 3) {
        while (breadcrumbs.length !== 4) {
          breadcrumbs.shift()
        }

        breadcrumbs[0].name = '...'
      }

      return breadcrumbs
    }
  },
  methods: {
    ...mapMutations([ 'resetSharedSelected' ]),
    base64: function (name) {
      return window.btoa(unescape(encodeURIComponent(name)))
    },
    fetchData: async function () {
      this.loaded = false
      this.notFound = false
      this.$store.commit('resetSharedSelected')
      this.$store.commit('sharedMultiple', false)
      try {
        this.file = await api.getHash(encodeURIComponent(this.$route.params.pathMatch))
        if (this.file.isDir) this.file.items = this.file.items.map((item, index) => {
          item.index = index
          item.url = `/share/${this.hash}/${this.path}/${encodeURIComponent(item.name)}`
          return item
        })
        this.loaded = true
      } catch (e) {
        this.notFound = true
      }
    },
    keyEvent (event) {
      // Esc!
      if (event.keyCode === 27) {
        // If we're on a listing, unselect all
        // files and folders.
        if (this.sharedSelectedCount > 0) {
          this.resetSharedSelected()
        }
      }
    },
    toggleMultipleSelection () {
      this.$store.commit('sharedMultiple', !this.shared.multiple)
    }
  }
}
</script>
