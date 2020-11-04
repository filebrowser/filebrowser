<template>
  <div class="share" v-if="loaded">
    <div class="share__box share__box__info">
        <div class="share__box__header">
          {{ file.isDir ? $t('download.downloadFolder') : $t('download.downloadFile') }}
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
    </div>
    <div v-if="file.isDir" class="share__box share__box__items">
      <div class="share__box__header" v-if="file.isDir">
        {{ $t('files.files') }}
      </div>
      <div id="listing" class="list">
        <div class="item" v-for="(item) in file.items.slice(0, this.showLimit)" :key="base64(item.name)">
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
    showLimit: 500
  }),
  watch: {
    '$route': 'fetchData'
  },
  created: function () {
    this.fetchData()
  },
  computed: {
    hash: function () {
      return this.$route.params.pathMatch
    },
    link: function () {
      return `${baseURL}/api/public/dl/${this.hash}/${encodeURI(this.file.name)}`
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
    }
  },
  methods: {
    base64: function (name) {
      return window.btoa(unescape(encodeURIComponent(name)))
    },
    fetchData: async function () {
      try {
        this.file = await api.getHash(this.hash)
        this.loaded = true
      } catch (e) {
        this.notFound = true
      }
    }
  }
}
</script>
