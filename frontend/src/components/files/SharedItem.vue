<template>
  <div class="item"
  role="button"
  tabindex="0"
  @click="click"
  @dblclick="dblclick"
  @touchstart="touchstart"
  :data-dir="isDir"
  :aria-label="name"
  :aria-selected="isSelected">
    <div>
      <i class="material-icons">{{ icon }}</i>
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

export default {
  name: 'sharedItem',
  data: function () {
    return {
      touches: 0
    }
  },
  props: ['name', 'isDir', 'url', 'type', 'size', 'modified', 'index'],
  computed: {
    ...mapState(['shared']),
    ...mapGetters(['sharedSelectedCount']),
    isSelected () {
      return (this.shared.selected.indexOf(this.index) !== -1)
    },
    icon () {
      if (this.isDir) return 'folder'
      if (this.type === 'image') return 'insert_photo'
      if (this.type === 'audio') return 'volume_up'
      if (this.type === 'video') return 'movie'
      return 'insert_drive_file'
    }
  },
  methods: {
    ...mapMutations(['addSharedSelected', 'removeSharedSelected', 'resetSharedSelected']),
    humanSize: function () {
      return filesize(this.size)
    },
    humanTime: function () {
      return moment(this.modified).fromNow()
    },
    click: function (event) {
      if (this.sharedSelectedCount !== 0) event.preventDefault()
      if (this.$store.state.shared.selected.indexOf(this.index) !== -1) {
        this.removeSharedSelected(this.index)
        return
      }

      if (event.shiftKey && this.shared.selected.length > 0) {
        let fi = 0
        let la = 0

        if (this.index > this.shared.selected[0]) {
          fi = this.shared.selected[0] + 1
          la = this.index
        } else {
          fi = this.index
          la = this.shared.selected[0] - 1
        }

        for (; fi <= la; fi++) {
          if (this.$store.state.shared.selected.indexOf(fi) == -1) {
            this.addSharedSelected(fi)
          }
        }

        return
      }

      if (!event.ctrlKey && !event.metaKey && !this.$store.state.shared.multiple) this.resetSharedSelected()
      this.addSharedSelected(this.index)
    },
    dblclick: function () {
      this.open()
    },
    touchstart () {
      setTimeout(() => {
        this.touches = 0
      }, 300)

      this.touches++
      if (this.touches > 1) {
        this.open()
      }
    },
    open: function () {
      this.$router.push({path: this.url})
    }
  }
}
</script>