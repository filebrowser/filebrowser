<template>
  <div class="prompt">
    <h3>File Information</h3>

    <p v-show="selected.length > 1">{{ selected.length }} files selected.</p>

    <p v-show="selected.length < 2"><strong>Display Name:</strong> {{ name() }}</p>
    <p><strong>Size:</strong> <span id="content_length"></span>{{ humanSize() }}</p>
    <p v-show="selected.length < 2"><strong>Last Modified:</strong> {{ humanTime() }}</p>

    <section v-show="dir() && selected.length === 0">
      <p><strong>Number of files:</strong> {{ req.numFiles }}</p>
      <p><strong>Number of directories:</strong> {{ req.numDirs }}</p>
    </section>

    <section v-show="!dir()">
      <p><strong>MD5:</strong> <code><a @click="checksum($event, 'md5')">show</a></code></p>
      <p><strong>SHA1:</strong> <code><a @click="checksum($event, 'sha1')">show</a></code></p>
      <p><strong>SHA256:</strong> <code><a @click="checksum($event, 'sha256')">show</a></code></p>
      <p><strong>SHA512:</strong> <code><a @click="checksum($event, 'sha512')">show</a></code></p>
    </section>

    <div>
      <button type="submit" @click="$store.commit('showInfo', false)" class="ok">OK</button>
    </div>
  </div>
</template>

<script>
import {mapState, mapGetters} from 'vuex'
import filesize from 'filesize'
import moment from 'moment'
import api from '@/utils/api'

export default {
  name: 'info',
  computed: {
    ...mapState(['req', 'selected']),
    ...mapGetters(['selectedCount'])
  },
  methods: {
    humanSize: function () {
      if (this.selectedCount === 0 || this.req.kind !== 'listing') {
        return filesize(this.req.size)
      }

      var sum = 0

      for (let i = 0; i < this.selectedCount; i++) {
        sum += this.req.items[this.selected[i]].size
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
      if (this.selectedCount === 0) {
        return this.req.name
      }

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
      event.preventDefault()

      let link

      if (this.selectedCount) {
        link = this.req.items[this.selected[0]].url
      } else {
        link = this.$route.path
      }

      api.checksum(link, hash)
        .then((hash) => {
          event.target.innerHTML = hash
        })
        .catch(error => {
          console.log(error)
        })
    }
  }
}
</script>
