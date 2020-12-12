<template>
  <div v-if="loaded">
    <div id="breadcrumbs">
      <router-link :to="'/share/' + hash + '/' + this.root" :aria-label="$t('files.home')" :title="$t('files.home')">
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
            {{ file.isDir ? hasSelected ? $t('download.downloadSelected') : $t('download.downloadFolder') : $t('download.downloadFile') }}
          </div>
          <div class="share__box__element share__box__center share__box__icon">
            <i class="material-icons">{{ file.isDir ? 'folder' : 'insert_drive_file'}}</i>
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
              <input type="checkbox" v-model="multiple">
              {{ $t('buttons.selectMultiple') }}
            </label>
          </div>
      </div>
      <div v-if="file.isDir" class="share__box share__box__items">
        <div class="share__box__header" v-if="file.isDir">
          {{ $t('files.files') }}
        </div>
        <div id="listing" class="list">
          <div class="item" v-for="(item) in file.items.slice(0, this.showLimit)" :key="base64(item.name)"
               :aria-selected="selected.includes(item.name)"
               @click="click(item.name)"
               @dblclick="dblclick(item.path)"
               @touchstart="touchstart(item.path)"
          >
            <div>
              <i class="material-icons">{{ item.isDir ? 'folder' : (item.type==='image') ? 'insert_photo' : 'insert_drive_file' }}</i>
            </div>
            <div>
              <p class="name">{{ item.name }}</p>
            </div>
          </div>
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
import { share as api } from '@/api'
import { baseURL } from '@/utils/constants'
import filesize from 'filesize'
import moment from 'moment'
import QrcodeVue from 'qrcode.vue'

export default {
  name: 'share',
  components: {
    QrcodeVue
  },
  data: () => ({
    loaded: false,
    notFound: false,
    file: null,
    showLimit: 500,
    multiple: false,
    touches: 0,
    selected: [],
    firstSelected: -1,
    root: ''
  }),
  watch: {
    '$route': 'fetchData'
  },
  created: async function () {
    await this.fetchData()
    this.root = this.file.path.split('/')[1]
  },
  mounted () {
    window.addEventListener('keydown', this.keyEvent)
  },
  beforeDestroy () {
    window.removeEventListener('keydown', this.keyEvent)
  },
  computed: {
    hasSelected: function () {
      return this.selected.length > 0
    },
    hash: function () {
      return this.$route.params.pathMatch.split('/')[0]
    },
    link: function () {
      let path = this.file.path.endsWith('/') ? this.file.path.slice(0, this.file.path.length - 1) : this.file.path
      if (!this.hasSelected) return `${baseURL}/api/public/dl/${this.hash}${path}`
      if (this.selected.length === 1) return `${baseURL}/api/public/dl/${this.hash}${path}/${encodeURI(this.selected[0])}`
      let files = []
      for (let s of this.selected) {
        files.push(encodeURI(s))
      }
      return `${baseURL}/api/public/dl/${this.hash}${path}/?files=${encodeURIComponent(files.join(','))}`
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
      let parts = this.file.path.split('/')

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
    base64: function (name) {
      return window.btoa(unescape(encodeURIComponent(name)))
    },
    fetchData: async function () {
      this.loaded = false
      this.notFound = false
      this.multiple = false
      this.touches = 0
      this.selected = []
      this.firstSelected = -1
      try {
        this.file = await api.getHash(this.$route.params.pathMatch)
        this.loaded = true
      } catch (e) {
        this.notFound = true
      }
    },
    fileItemsIndexOf: function (name) {
      return this.file.items.indexOf(this.file.items.filter(item => item.name === name)[0])
    },
    addSelected: function(name) {
      this.selected.push(name)
    },
    removeSelected: function (name) {
      let i = this.selected.indexOf(name)
      if (i === -1) return
      this.selected.splice(i, 1)
      if (i === 0 && this.hasSelected) {
        this.firstSelected = this.fileItemsIndexOf(this.selected[0])
      }
    },
    resetSelected: function () {
      this.selected = []
      this.firstSelected = -1
    },
    click: function (name) {
      if (this.hasSelected) event.preventDefault()
      if (this.selected.indexOf(name) !== -1) {
        this.removeSelected(name)
        return
      }

      let index = this.fileItemsIndexOf(name)
      if (event.shiftKey && this.hasSelected) {
        let fi = 0
        let la = 0

        if (index > this.firstSelected) {
          fi = this.firstSelected + 1
          la = index
        } else {
          fi = index
          la = this.firstSelected - 1
        }

        for (; fi <= la; fi++) {
          if (this.selected.indexOf(this.file.items[fi].name) === -1) {
            this.addSelected(this.file.items[fi].name)
          }
        }

        return
      }

      if (!event.ctrlKey && !event.metaKey && !this.multiple) this.resetSelected()
      if (this.firstSelected === -1) this.firstSelected = index
      this.addSelected(name)
    },
    dblclick: function (path) {
      this.$router.push({path: `/share/${this.hash}${path}`})
    },
    touchstart (path) {
      setTimeout(() => {
        this.touches = 0
      }, 300)

      this.touches++
      if (this.touches > 1) {
        this.dblclick(path)
      }
    },
    keyEvent (event) {
      // Esc!
      if (event.keyCode === 27) {
        // If we're on a listing, unselect all
        // files and folders.
        if (this.hasSelected) {
          this.resetSelected()
        }
      }
    }
  }
}
</script>
