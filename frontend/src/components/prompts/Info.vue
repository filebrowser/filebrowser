<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t('prompts.fileInfo') }}</h2>
    </div>

    <div class="card-content">
      <p v-if="selected.length > 1">{{ $t('prompts.filesSelected', { count: selected.length }) }}</p>

      <p class="break-word" v-if="selected.length < 2"><strong>{{ $t('prompts.displayName') }}</strong> {{ name }}</p>
      <p v-if="!dir || selected.length > 1"><strong>{{ $t('prompts.size') }}:</strong> <span id="content_length"></span> {{ humanSize }}</p>
      <p v-if="selected.length < 2"><strong>{{ $t('prompts.lastModified') }}:</strong> {{ humanTime }}</p>

      <template v-if="dir && selected.length === 0">
        <p><strong>{{ $t('prompts.numberFiles') }}:</strong> {{ req.numFiles }}</p>
        <p><strong>{{ $t('prompts.numberDirs') }}:</strong> {{ req.numDirs }}</p>
      </template>

      <template v-if="!dir">
        <p><strong>MD5: </strong><code><a @click="checksum($event, 'md5')">{{ $t('prompts.show') }}</a></code></p>
        <p><strong>SHA1: </strong><code><a @click="checksum($event, 'sha1')">{{ $t('prompts.show') }}</a></code></p>
        <p><strong>SHA256: </strong><code><a @click="checksum($event, 'sha256')">{{ $t('prompts.show') }}</a></code></p>
        <p><strong>SHA512: </strong><code><a @click="checksum($event, 'sha512')">{{ $t('prompts.show') }}</a></code></p>
      </template>
    </div>

    <div class="card-action">
      <button type="submit"
        @click="$store.commit('closeHovers')"
        class="button button--flat"
        :aria-label="$t('buttons.ok')"
        :title="$t('buttons.ok')">{{ $t('buttons.ok') }}</button>
    </div>
  </div>
</template>

<script>
import {mapState, mapGetters} from 'vuex'
import filesize from 'filesize'
import moment from 'moment'
import { files as api } from '@/api'

export default {
  name: 'info',
  computed: {
    ...mapState(['req', 'selected']),
    ...mapGetters(['selectedCount', 'isListing']),
    humanSize: function () {
      if (this.selectedCount === 0 || !this.isListing) {
        return filesize(this.req.size)
      }

      let sum = 0

      for (let selected of this.selected) {
        sum += this.req.items[selected].size
      }

      return filesize(sum)
    },
    humanTime: function () {
      if (this.selectedCount === 0) {
        return moment(this.req.modified).fromNow()
      }

      return moment(this.req.items[this.selected[0]]).fromNow()
    },
    name: function () {
      return this.selectedCount === 0 ? this.req.name : this.req.items[this.selected[0]].name
    },
    dir: function () {
      return this.selectedCount > 1 || (this.selectedCount === 0
        ? this.req.isDir
        : this.req.items[this.selected[0]].isDir)
    }
  },
  methods: {
    checksum: async function (event, algo) {
      event.preventDefault()

      let link

      if (this.selectedCount) {
        link = this.req.items[this.selected[0]].url
      } else {
        link = this.$route.path
      }

      try {
        const hash = await api.checksum(link, algo)
        // eslint-disable-next-line
        event.target.innerHTML = hash
      } catch (e) {
        this.$showError(e)
      }
    }
  }
}
</script>
