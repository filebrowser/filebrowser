<template>
    <div class="prompt">
        <h3>File Information</h3>

        <p v-show="selected.length > 1">{{ selected.length }} files selected.</p>

        <p v-show="selected.length < 2"><strong>Display Name:</strong> {{ name() }}</p>
        <p><strong>Size:</strong> <span id="content_length"></span>{{ humanSize() }}</p>
        <p v-show="selected.length < 2"><strong>Last Modified:</strong> {{ humanTime() }}</p>

        <section v-show="dir() && selected.length === 0">
          <p><strong>Number of files:</strong> {{ req.data.numFiles }}</p>
          <p><strong>Number of directories:</strong> {{ req.data.numDirs }}</p>
        </section>

        <section v-show="!dir()">
            <p><strong>MD5:</strong> <code><a @click="checksum($event, 'md5')">show</a></code></p>
            <p><strong>SHA1:</strong> <code><a @click="checksum($event, 'sha1')">show</a></code></p>
            <p><strong>SHA256:</strong> <code><a @click="checksum($event, 'sha256')">show</a></code></p>
            <p><strong>SHA512:</strong> <code><a @click="checksum($event, 'sha512')">show</a></code></p>
        </section>

        <div>
            <button type="submit" @click="close" class="ok">OK</button>
        </div>
    </div>
</template>

<script>
import {mapState} from 'vuex'
import filesize from 'filesize'
import moment from 'moment'

export default {
  name: 'info-prompt',
  data: function () {
    return window.info
  },
  computed: mapState(['req', 'selected']),
  methods: {
    humanSize: function () {
      if (this.selected.length === 0 || this.req.kind !== 'listing') {
        return filesize(this.req.data.size)
      }

      var sum = 0

      for (let i = 0; i < this.selected.length; i++) {
        sum += this.req.data.items[this.selected[i]].size
      }

      return filesize(sum)
    },
    humanTime: function () {
      if (this.selected.length === 0) {
        return moment(this.req.data.modified).fromNow()
      }

      return moment(this.req.data.items[this.selected[0]]).fromNow()
    },
    name: function () {
      if (this.selected.length === 0) {
        return this.req.data.name
      }

      return this.req.data.items[this.selected[0]].name
    },
    dir: function () {
      if (this.selected.length > 1) {
        // Don't show when multiple selected.
        return true
      }

      if (this.selected.length === 0) {
        return this.req.data.isDir
      }

      return this.req.data.items[this.selected[0]].isDir
    },
    checksum: function (event, hash) {
      event.preventDefault()

      let request = new window.XMLHttpRequest()
      let link

      if (this.selected.length) {
        link = this.req.data.items[this.selected[0]].url
      } else {
        link = window.location.pathname
      }

      request.open('GET', `${link}?checksum=${hash}`, true)

      request.onload = () => {
        if (request.status >= 300) {
          console.log(request.statusText)
          return
        }

        event.target.innerHTML = request.responseText
      }

      request.onerror = (e) => console.log(e)
      request.send()
    },
    close: function () {
      this.$store.commit('showInfo', false)
    }
  }
}
</script>
