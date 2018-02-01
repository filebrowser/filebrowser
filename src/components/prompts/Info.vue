<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t('prompts.fileInfo') }}</h2>
    </div>

    <div class="card-content">
      <p v-if="selected.length > 1">{{ $t('prompts.filesSelected', { count: selected.length }) }}</p>

      <p v-if="selected.length < 2"><strong>{{ $t('prompts.displayName') }}</strong> {{ name() }}</p>
      <p><strong>{{ $t('prompts.size') }}:</strong> <span id="content_length"></span>{{ humanSize() }}</p>
      <p v-if="selected.length < 2"><strong>{{ $t('prompts.lastModified') }}:</strong> {{ humanTime() }}</p>

      <template v-if="dir() && selected.length === 0">
        <p><strong>{{ $t('prompts.numberFiles') }}:</strong> {{ req.numFiles }}</p>
        <p><strong>{{ $t('prompts.numberDirs') }}:</strong> {{ req.numDirs }}</p>
      </template>

      <template v-if="!dir()">
        <p><strong>MD5:</strong> <code><a @click="checksum($event, 'md5')">{{ $t('prompts.show') }}</a></code></p>
        <p><strong>SHA1:</strong> <code><a @click="checksum($event, 'sha1')">{{ $t('prompts.show') }}</a></code></p>
        <p><strong>SHA256:</strong> <code><a @click="checksum($event, 'sha256')">{{ $t('prompts.show') }}</a></code></p>
        <p><strong>SHA512:</strong> <code><a @click="checksum($event, 'sha512')">{{ $t('prompts.show') }}</a></code></p>
      </template>
    </div>

    <div class="card-action">
      <button type="submit"
        @click="$store.commit('closeHovers')"
        class="flat"
        :aria-label="$t('buttons.ok')"
        :title="$t('buttons.ok')">{{ $t('buttons.ok') }}</button>
    </div>
  </div>
</template>

<script>
import {mapState, mapGetters} from 'vuex'
import filesize from 'filesize'
import moment from 'moment'
import * as api from '@/utils/api'

export default {
  name: 'info',
  computed: {
    ...mapState(['req', 'selected']),
    ...mapGetters(['selectedCount'])
  },
  methods: {
    humanSize: function () {
      // If there are no files selected or this is not a listing
      // show the human file size of the current request.
      if (this.selectedCount === 0 || this.req.kind !== 'listing') {
        return filesize(this.req.size)
      }

      // Otherwise, sum the sizes of each selected file and returns
      // its human form.
      var sum = 0

      for (let i = 0; i < this.selectedCount; i++) {
        sum += this.req.items[this.selected[i]].size
      }

      return filesize(sum)
    },
    humanTime: function () {
      // If there are no selected files, return the current request
      // modified time.
      if (this.selectedCount === 0) {
        return moment(this.req.modified).fromNow()
      }

      // Otherwise return the modified time of the first item
      // that is selected since this should not appear when
      // there is more than one file selected.
      return moment(this.req.items[this.selected[0]]).fromNow()
    },
    name: function () {
      // Return the name of the current opened file if there
      // are no selected files.
      if (this.selectedCount === 0) {
        return this.req.name
      }

      // Otherwise, just return the name of the selected file.
      // This field won't show when there is more than one
      // file selected.
      return this.req.items[this.selected[0]].name
    },
    dir: function () {
      if (this.selectedCount > 1) {
        // Don't show when multiple selected.
        return true
      }

      if (this.selectedCount === 0) {
        return this.req.isDir
      }

      return this.req.items[this.selected[0]].isDir
    },
    checksum: function (event, hash) {
      // Gets the checksum of the current selected or
      // opened file. Doesn't work for directories.
      event.preventDefault()

      let link

      if (this.selectedCount) {
        link = this.req.items[this.selected[0]].url
      } else {
        link = this.$route.path
      }

      api.checksum(link, hash)
        .then((hash) => { event.target.innerHTML = hash })
        .catch(this.$showError)
    }
  }
}
</script>
