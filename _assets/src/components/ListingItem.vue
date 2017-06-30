<template>
    <div class="item"
        draggable="true"
        @dragstart="dragStart"
        @dragover="dragOver"
        @drop="drop"
        @click="click"
        @dblclick="open"
        :id="base64()">
        <div>
            <i class="material-icons">{{ icon() }}</i>
        </div>

        <div>
            <p class="name">{{ name }}</p>

            <p v-if="isDir" class="size" data-order="-1">&mdash;</p>
            <p v-else class="size" :data-order="humanSize()">{{ humanSize() }}</p>

            <p class="modified">
                <time :datetime="modified">{{ humanTime() }}</time>
            </p>
        </div>
    </div>
</template>

<script>
import { mapMutations, mapGetters } from 'vuex'
import filesize from 'filesize'
import moment from 'moment'
import webdav from '../utils/webdav.js'
import page from '../utils/page.js'

export default {
  name: 'item',
  props: ['name', 'isDir', 'url', 'type', 'size', 'modified', 'index'],
  methods: {
    ...mapGetters(['selectedCount']),
    ...mapMutations(['addSelected', 'removeSelected']),
    icon: function () {
      if (this.isDir) return 'folder'
      if (this.type === 'image') return 'insert_photo'
      if (this.type === 'audio') return 'volume_up'
      if (this.type === 'video') return 'movie'
      return 'insert_drive_file'
    },
    humanSize: function () {
      return filesize(this.size)
    },
    humanTime: function () {
      return moment(this.modified).fromNow()
    },
    base64: function () {
      return window.btoa(this.name)
    },
    dragStart: function (event) {
      let el = event.target

      for (let i = 0; i < 5; i++) {
        if (!el.classList.contains('item')) {
          el = el.parentElement
        }
      }

      event.dataTransfer.setData('name', el.querySelector('.name').innerHTML)
      event.dataTransfer.setData('obj-url', this.url)
    },
    dragOver: function (event) {
      if (!this.isDir) return

      event.preventDefault()
      let el = event.target

      for (let i = 0; i < 5; i++) {
        if (!el.classList.contains('item')) {
          el = el.parentElement
        }
      }

      el.style.opacity = 1
    },
    drop: function (event) {
      if (!this.isDir) return
      event.preventDefault()

      let url = event.dataTransfer.getData('obj-url')
      let name = event.dataTransfer.getData('name')

      if (name === '' || url === '' || url === this.url) return

      webdav.move(url, this.url + name)
        .then(() => page.reload())
        .catch(error => console.log(error))
    },
    unselectAll: function () {
      let items = document.getElementsByClassName('item')
      Array.from(items).forEach(link => {
        link.setAttribute('aria-selected', false)
      })

      this.$store.commit('resetSelected')
      return false
    },
    click: function (event) {
      if (this.selectedCount !== 0) event.preventDefault()
      if (this.$store.state.selected.indexOf(this.index) === -1) {
        if (!event.ctrlKey && !this.$store.state.multiple) this.unselectAll()

        this.$el.setAttribute('aria-selected', true)
        // WORKS: this.$store.commit('addSelected', this.index)
        this.addSelected(this.index)
      } else {
        this.$el.setAttribute('aria-selected', false)
        this.removeSelected(this.index)
        // WORKS: this.$store.commit('removeSelected', this.index)
      }

      // this.handleSelectionChange()
      return false
    },
    open: function (event) {
      page.open(this.url)
    }
  }
}
</script>
