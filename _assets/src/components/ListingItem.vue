<template>
  <div class="item"
  draggable="true"
  @dragstart="dragStart"
  @dragover="dragOver"
  @drop="drop"
  @click="click"
  @dblclick="open"
  :aria-selected="isSelected()">
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
import { mapMutations, mapGetters, mapState } from 'vuex'
import filesize from 'filesize'
import moment from 'moment'
import webdav from '../utils/webdav.js'
import page from '../utils/page.js'

export default {
  name: 'item',
  props: ['name', 'isDir', 'url', 'type', 'size', 'modified', 'index'],
  computed: {
    ...mapState(['selected', 'req']),
    ...mapGetters(['selectedCount'])
  },
  methods: {
    ...mapMutations(['addSelected', 'removeSelected', 'resetSelected']),
    isSelected: function () {
      return (this.selected.indexOf(this.index) !== -1)
    },
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
    dragStart: function (event) {
      if (this.selectedCount === 0) {
        this.addSelected(this.index)
      }
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

      if (this.selectedCount === 0) return

      let promises = []

      for (let i of this.selected) {
        let url = this.req.data.items[i].url
        let name = this.req.data.items[i].name

        promises.push(webdav.move(url, this.url + name))
      }

      Promise.all(promises)
        .then(() => page.reload())
        .catch(error => console.log(error))
    },
    click: function (event) {
      if (this.selectedCount !== 0) event.preventDefault()
      if (this.$store.state.selected.indexOf(this.index) === -1) {
        if (!event.ctrlKey && !this.$store.state.multiple) this.resetSelected()

        this.addSelected(this.index)
      } else {
        this.removeSelected(this.index)
      }

      return false
    },
    open: function (event) {
      page.open(this.url)
    }
  }
}
</script>
