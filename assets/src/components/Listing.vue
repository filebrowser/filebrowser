<template>
  <div v-if="(req.numDirs + req.numFiles) == 0">
    <h2 class="message">
      <i class="material-icons">sentiment_dissatisfied</i>
      <span>It feels lonely here...</span>
    </h2>
    <input style="display:none" type="file" id="upload-input" @change="uploadInput($event)" value="Upload" multiple>
  </div>
  <div v-else id="listing"
    :class="req.display"
    @drop="drop"
    @dragenter="dragEnter"
    @dragend="dragEnd">
    <div>
      <div class="item header">
        <div></div>
        <div>
          <p :class="{ active: nameSorted }" class="name" @click="sort('name')">
            <span>Name</span>
            <i class="material-icons">{{ nameIcon }}</i>
          </p>

          <p :class="{ active: !nameSorted }" class="size" @click="sort('size')">
            <span>Size</span>
            <i class="material-icons">{{ sizeIcon }}</i>
          </p>

          <p class="modified">Last modified</p>
        </div>
      </div>
    </div>

    <h2 v-if="req.numDirs > 0">Folders</h2>
    <div v-if="req.numDirs > 0">
      <item v-for="(item, index) in req.items"
        v-if="item.isDir"
        :key="base64(item.name)"
        v-bind:index="index"
        v-bind:name="item.name"
        v-bind:isDir="item.isDir"
        v-bind:url="item.url"
        v-bind:modified="item.modified"
        v-bind:type="item.type"
        v-bind:size="item.size">
      </item>
    </div>

    <h2 v-if="req.numFiles > 0">Files</h2>
    <div v-if="req.numFiles > 0">
      <item v-for="(item, index) in req.items"
        v-if="!item.isDir"
        :key="base64(item.name)"
        v-bind:index="index"
        v-bind:name="item.name"
        v-bind:isDir="item.isDir"
        v-bind:url="item.url"
        v-bind:modified="item.modified"
        v-bind:type="item.type"
        v-bind:size="item.size">
      </item>
    </div>

    <input style="display:none" type="file" id="upload-input" @change="uploadInput($event)" value="Upload" multiple>

    <div v-show="$store.state.multiple" :class="{ active: $store.state.multiple }" id="multiple-selection">
    <p>Multiple selection enabled</p>
      <div @click="$store.commit('multiple', false)" tabindex="0" role="button" title="Clear" aria-label="Clear" class="action">
        <i class="material-icons" title="Clear">clear</i>
      </div>
    </div>
  </div>
</template>

<script>
import {mapState} from 'vuex'
import Item from './ListingItem'
import css from '@/utils/css'
import api from '@/utils/api'
import buttons from '@/utils/buttons'

export default {
  name: 'listing',
  components: { Item },
  computed: {
    ...mapState(['req', 'selected']),
    nameSorted () {
      return (this.req.sort === 'name')
    },
    ascOrdered () {
      return (this.req.order === 'asc')
    },
    nameIcon () {
      if (this.nameSorted && !this.ascOrdered) {
        return 'arrow_upward'
      }

      return 'arrow_downward'
    },
    sizeIcon () {
      if (!this.nameSorted && this.ascOrdered) {
        return 'arrow_downward'
      }

      return 'arrow_upward'
    }
  },
  mounted: function () {
    // Check the columns size for the first time.
    this.resizeEvent()

    // Add the needed event listeners to the window and document.
    window.addEventListener('keydown', this.keyEvent)
    window.addEventListener('resize', this.resizeEvent)
    document.addEventListener('dragover', this.preventDefault)
    document.addEventListener('drop', this.drop)
  },
  beforeDestroy () {
    // Remove event listeners before destroying this page.
    window.removeEventListener('keydown', this.keyEvent)
    window.removeEventListener('resize', this.resizeEvent)
    document.removeEventListener('dragover', this.preventDefault)
    document.removeEventListener('drop', this.drop)
  },
  methods: {
    base64: function (name) {
      return window.btoa(unescape(encodeURIComponent(name)))
    },
    keyEvent (event) {
      if (!event.ctrlKey && !event.metaKey) {
        return
      }

      let key = String.fromCharCode(event.which).toLowerCase()

      switch (key) {
        case 'f':
          event.preventDefault()
          this.$store.commit('showHover', 'search')
          break
        case 'c':
        case 'x':
          this.copyCut(event, key)
          break
        case 'v':
          this.paste(event)
          break
      }
    },
    preventDefault (event) {
      // Wrapper around prevent default.
      event.preventDefault()
    },
    copyCut (event, key) {
      event.preventDefault()
      let items = []

      for (let i of this.selected) {
        items.push({
          from: this.req.items[i].url,
          name: encodeURIComponent(this.req.items[i].name)
        })
      }

      this.$store.commit('updateClipboard', {
        key: key,
        items: items
      })
    },
    paste (event) {
      event.preventDefault()

      let items = []

      for (let item of this.$store.state.clipboard.items) {
        items.push({
          from: item.from,
          to: this.$route.path + item.name
        })
      }

      if (this.$store.state.clipboard.key === 'x') {
        api.move(items).then(() => {
          this.$store.commit('setReload', true)
        }).catch(error => {
          this.$store.commit('showError', error)
        })
        return
      }

      api.copy(items).then(() => {
        this.$store.commit('setReload', true)
      }).catch(error => {
        this.$store.commit('showError', error)
      })
    },
    resizeEvent () {
      // Update the columns size based on the window width.
      let columns = Math.floor(document.querySelector('main').offsetWidth / 300)
      let items = css(['#listing.mosaic .item', '.mosaic#listing .item'])
      if (columns === 0) columns = 1
      items.style.width = `calc(${100 / columns}% - 1em)`
    },
    dragEnter: function (event) {
      // When the user starts dragging an item, put every
      // file on the listing with 50% opacity.
      let items = document.getElementsByClassName('item')

      Array.from(items).forEach(file => {
        file.style.opacity = 0.5
      })
    },
    dragEnd: function (event) {
      this.resetOpacity()
    },
    drop: function (event) {
      event.preventDefault()

      let dt = event.dataTransfer
      let files = dt.files
      let el = event.target

      for (let i = 0; i < 5; i++) {
        if (el !== null && !el.classList.contains('item')) {
          el = el.parentElement
        }
      }

      if (files.length > 0) {
        if (el !== null && el.classList.contains('item') && el.dataset.dir === 'true') {
          this.handleFiles(files, el.querySelector('.name').innerHTML + '/')
          return
        }

        this.handleFiles(files, '')
      } else {
        this.resetOpacity()
      }
    },
    uploadInput: function (event) {
      this.handleFiles(event.currentTarget.files, '')
    },
    resetOpacity: function () {
      let items = document.getElementsByClassName('item')

      Array.from(items).forEach(file => {
        file.style.opacity = 1
      })
    },
    handleFiles: function (files, base) {
      this.resetOpacity()

      buttons.loading('upload')
      let promises = []

      for (let file of files) {
        promises.push(api.post(this.$route.path + base + file.name, file))
      }

      Promise.all(promises)
        .then(() => {
          buttons.done('upload')
          this.$store.commit('setReload', true)
        })
        .catch(error => {
          buttons.done('upload')
          this.$store.commit('showError', error)
        })

      return false
    },
    sort (sort) {
      let order = 'desc'

      if (sort === 'name') {
        if (this.nameIcon === 'arrow_upward') {
          order = 'asc'
        }
      } else {
        if (this.sizeIcon === 'arrow_upward') {
          order = 'asc'
        }
      }

      let path = this.$store.state.baseURL
      if (path === '') path = '/'
      document.cookie = `sort=${sort}; max-age=31536000; path=${path}`
      document.cookie = `order=${order}; max-age=31536000; path=${path}`
      this.$store.commit('setReload', true)
    }
  }
}
</script>
