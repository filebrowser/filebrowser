<template>
  <div v-if="loaded" class="share">
    <a target="_blank" :href="link">
      <div class="share__box">
        <div v-if="file.isDir" class="share__box__download">{{ $t('download.downloadFolder') }}</div>
        <div v-else class="share__box__download">{{ $t('download.downloadFile') }}</div>
        <div class="share__box__info">
          <svg v-if="file.isDir" fill="#40c4ff" height="150" viewBox="0 0 24 24" width="150" xmlns="http://www.w3.org/2000/svg">
            <path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z" />
            <path d="M0 0h24v24H0z" fill="none" />
          </svg>
          <svg v-else fill="#40c4ff" height="150" viewBox="0 0 24 24" width="150" xmlns="http://www.w3.org/2000/svg">
            <path d="M6 2c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6H6zm7 7V3.5L18.5 9H13z" />
            <path d="M0 0h24v24H0z" fill="none" />
          </svg>
          <h1 class="share__box__title">{{ file.name }}</h1>
          <qrcode-vue :value="fullLink" size="200" level="M" />
        </div>
      </div>
    </a>
  </div>
</template>

<script>
import { share as api } from '@/api'
import { baseURL } from '@/utils/constants'
import QrcodeVue from 'qrcode.vue'

export default {
  name: 'Share',
  components: {
    QrcodeVue
  },
  data: () => ({
    loaded: false,
    notFound: false,
    file: null
  }),
  computed: {
    hash: function() {
      return this.$route.params.pathMatch
    },
    link: function() {
      return `${baseURL}/api/public/dl/${this.hash}/${encodeURI(this.file.name)}`
    },
    fullLink: function() {
      return window.location.origin + this.link
    }
  },
  watch: {
    '$route': 'fetchData'
  },
  created: function() {
    this.fetchData()
  },
  methods: {
    fetchData: async function() {
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
