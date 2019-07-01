<template>
  <div class="share" v-if="loaded">
    <a target="_blank" :href="link" v-if="!file.isDir">
      <div class="share__box">
        <div class="share__box__download">{{ $t('download.downloadFile') }}</div>
        <div class="share__box__info">
          <svg
            fill="#40c4ff"
            height="150"
            viewBox="0 0 24 24"
            width="150"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M6 2c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6H6zm7 7V3.5L18.5 9H13z"
            />
            <path d="M0 0h24v24H0z" fill="none" />
          </svg>
          <h1 class="share__box__title">{{ file.name }}</h1>
          <qrcode-vue :value="fullLink" size="200" level="M"></qrcode-vue>
        </div>
      </div>
    </a>
    <listing></listing>
  </div>
</template>

<script>
import Listing from '@/components/files/Listing'

import { share as api } from '@/api'
import { baseURL } from '@/utils/constants'
import QrcodeVue from 'qrcode.vue'

export default {
  name: 'share',
  components: {
    QrcodeVue,
    Listing
  },
  data: () => ({
    loaded: false,
    notFound: false,
    file: null,
    dirContent: null
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
  },
  methods: {
    fetchData: async function () {
      try {
        this.file = await api.getHash(this.hash)

        if (this.file.isDir) {
          const tmpPath = this.hash.split('/')
          const hash = tmpPath.shift();
          const relativePath = tmpPath.join('/');

          const items = await fetch('/api/public/resources/' + hash, {
            headers: {
              'Content-Type': 'application/json',
              'Relative-Path': relativePath
            }
          }).then((ret) => {
            if (ret.status === 200) {
              return ret.json();
            }
          }).then((file) => {
            return file.items
          });
          console.log(items);

          this.dirContent = items;
        }

        this.loaded = true
      } catch (e) {
        this.notFound = true
      }
    }
  }
}
</script>
