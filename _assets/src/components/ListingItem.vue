<template>
    <div class="item" 
        draggable="true"
        :id="base64()"
        :data-dir="isDir" 
        :data-url="url" >
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
import filesize from 'filesize'
import moment from 'moment'

export default {
  name: 'item',
  props: ['name', 'isDir', 'url', 'type', 'size', 'modified'],
  methods: {
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
    }
  }
}
</script>
