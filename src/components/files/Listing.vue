<template>
  <div v-if="(req.numDirs + req.numFiles) == 0">
    <h2 class="message">
      <i class="material-icons">sentiment_dissatisfied</i>
      <span>{{ $t('files.lonely') }}</span>
    </h2>
    <input style="display:none" type="file" id="upload-input" @change="uploadInput($event)" multiple>
  </div>
  <div v-else id="listing"
    :class="user.viewMode"
    @dragenter="dragEnter"
    @dragend="dragEnd">
    <div>
      <div class="item header">
        <div></div>
        <div>
          <p :class="{ active: nameSorted }" class="name"
            role="button"
            tabindex="0"
            @click="sort('name')"
            :title="$t('files.sortByName')"
            :aria-label="$t('files.sortByName')">
            <span>{{ $t('files.name') }}</span>
            <i class="material-icons">{{ nameIcon }}</i>
          </p>

          <p :class="{ active: sizeSorted }" class="size"
            role="button"
            tabindex="0"
            @click="sort('size')"
            :title="$t('files.sortBySize')"
            :aria-label="$t('files.sortBySize')">
            <span>{{ $t('files.size') }}</span>
            <i class="material-icons">{{ sizeIcon }}</i>
          </p>
          <p :class="{ active: modifiedSorted }" class="modified"
            role="button"
            tabindex="0"
            @click="sort('modified')"
            :title="$t('files.sortByLastModified')"
            :aria-label="$t('files.sortByLastModified')">
            <span>{{ $t('files.lastModified') }}</span>
            <i class="material-icons">{{ modifiedIcon }}</i>
          </p>
        </div>
      </div>
    </div>

    <h2 v-if="req.numDirs > 0">{{ $t('files.folders') }}</h2>
    <div v-if="req.numDirs > 0">
      <item v-for="(item) in dirs"
        :key="base64(item.name)"
        v-bind:index="item.index"
        v-bind:name="item.name"
        v-bind:isDir="item.isDir"
        v-bind:url="item.url"
        v-bind:modified="item.modified"
        v-bind:type="item.type"
        v-bind:size="item.size">
      </item>
    </div>

    <h2 v-if="req.numFiles > 0">{{ $t('files.files') }}</h2>
    <div v-if="req.numFiles > 0">
      <item v-for="(item) in files"
        :key="base64(item.name)"
        v-bind:index="item.index"
        v-bind:name="item.name"
        v-bind:isDir="item.isDir"
        v-bind:url="item.url"
        v-bind:modified="item.modified"
        v-bind:type="item.type"
        v-bind:size="item.size">
      </item>
    </div>

    <input style="display:none" type="file" id="upload-input" @change="uploadInput($event)" multiple>

    <div :class="{ active: $store.state.multiple }" id="multiple-selection">
    <p>{{ $t('files.multipleSelectionEnabled') }}</p>
      <div @click="$store.commit('multiple', false)" tabindex="0" role="button" :title="$t('files.clear')" :aria-label="$t('files.clear')" class="action">
        <i class="material-icons">clear</i>
      </div>
    </div>
  </div>
</template>

<script>
import { mapState, mapMutations } from 'vuex'
import Item from './ListingItem'
import css from '@/utils/css'
import { users, files as api } from '@/api'
import buttons from '@/utils/buttons'

export default {
  name: 'listing',
  components: { Item },
  data: function () {
    return {
      show: 50
    }
  },
  computed: {
    ...mapState(['req', 'selected', 'user']),
    nameSorted () {
      return (this.req.sorting.by === 'name')
    },
    sizeSorted () {
      return (this.req.sorting.by === 'size')
    },
    modifiedSorted () {
      return (this.req.sorting.by === 'modified')
    },
    ascOrdered () {
      return this.req.sorting.asc
    },
    items () {
      const dirs = []
      const files = []

      this.req.items.forEach((item) => {
        if (item.isDir) {
          dirs.push(item)
        } else {
          files.push(item)
        }
      })

      return { dirs, files }
    },
    dirs () {
      return this.items.dirs.slice(0, this.show)
    },
    files () {
      let show = this.show - this.items.dirs.length

      if (show < 0) show = 0

      return this.items.files.slice(0, show)
    },
    nameIcon () {
      if (this.nameSorted && !this.ascOrdered) {
        return 'arrow_upward'
      }

      return 'arrow_downward'
    },
    sizeIcon () {
      if (this.sizeSorted && this.ascOrdered) {
        return 'arrow_downward'
      }

      return 'arrow_upward'
    },
    modifiedIcon () {
      if (this.modifiedSorted && this.ascOrdered) {
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
    window.addEventListener('scroll', this.scrollEvent)
    document.addEventListener('dragover', this.preventDefault)
    document.addEventListener('drop', this.drop)
  },
  beforeDestroy () {
    // Remove event listeners before destroying this page.
    window.removeEventListener('keydown', this.keyEvent)
    window.removeEventListener('resize', this.resizeEvent)
    window.removeEventListener('scroll', this.scrollEvent)
    document.removeEventListener('dragover', this.preventDefault)
    document.removeEventListener('drop', this.drop)
  },
  methods: {
    ...mapMutations([ 'updateUser' ]),
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
      if (event.target.tagName.toLowerCase() === 'input') {
        return
      }

      let items = []

      for (let i of this.selected) {
        items.push({
          from: this.req.items[i].url,
          name: encodeURIComponent(this.req.items[i].name)
        })
      }

      if (items.length == 0) {
        return
      }

      this.$store.commit('updateClipboard', {
        key: key,
        items: items
      })
    },
    paste (event) {
      if (event.target.tagName.toLowerCase() === 'input') {
        return
      }

      let items = []

      for (let item of this.$store.state.clipboard.items) {
        const from = item.from.endsWith('/') ? item.from.slice(0, -1) : item.from
        const to = this.$route.path + item.name
        items.push({ from, to })
      }

      if (items.length === 0) {
        return
      }

      if (this.$store.state.clipboard.key === 'x') {
        api.move(items).then(() => {
          this.$store.commit('setReload', true)
        }).catch(this.$showError)
        return
      }

      api.copy(items).then(() => {
        this.$store.commit('setReload', true)
      }).catch(this.$showError)
    },
    resizeEvent () {
      // Update the columns size based on the window width.
      let columns = Math.floor(document.querySelector('main').offsetWidth / 300)
      let items = css(['#listing.mosaic .item', '.mosaic#listing .item'])
      if (columns === 0) columns = 1
      items.style.width = `calc(${100 / columns}% - 1em)`
    },
    scrollEvent () {
      if ((window.innerHeight + window.scrollY) >= document.body.offsetHeight) {
        this.show += 50
      }
    },
    dragEnter () {
      // When the user starts dragging an item, put every
      // file on the listing with 50% opacity.
      let items = document.getElementsByClassName('item')

      Array.from(items).forEach(file => {
        file.style.opacity = 0.5
      })
    },
    dragEnd () {
      this.resetOpacity()
    },
    drop: function (event) {
      event.preventDefault()
      this.resetOpacity()

      let dt = event.dataTransfer
      let files = dt.files
      let el = event.target

      if (files.length <= 0) return

      for (let i = 0; i < 5; i++) {
        if (el !== null && !el.classList.contains('item')) {
          el = el.parentElement
        }
      }

      let base = ''
      if (el !== null && el.classList.contains('item') && el.dataset.dir === 'true') {
        base = el.querySelector('.name').innerHTML + '/'
      }

      if (base !== '') {
        api.fetch(this.$route.path + base)
          .then(req => {
            this.checkConflict(files, req.items, base)
          })
          .catch(this.$showError)

        return
      }

      this.checkConflict(files, this.req.items, base)
    },
    checkConflict (files, items, base) {
      if (typeof items === 'undefined' || items === null) {
        items = []
      }

      let conflict = false
      for (let i = 0; i < files.length; i++) {
        let res = items.findIndex(function hasConflict (element) {
          return (element.name === this)
        }, files[i].name)

        if (res >= 0) {
          conflict = true
          break
        }
      }

      if (!conflict) {
        this.handleFiles(files, base)
        return
      }

      this.$store.commit('showHover', {
        prompt: 'replace',
        confirm: (event) => {
          event.preventDefault()
          this.$store.commit('closeHovers')
          this.handleFiles(files, base, true)
        }
      })
    },
    uploadInput (event) {
      this.checkConflict(event.currentTarget.files, this.req.items, '')
    },
    resetOpacity () {
      let items = document.getElementsByClassName('item')

      Array.from(items).forEach(file => {
        file.style.opacity = 1
      })
    },
    handleFiles (files, base, overwrite = false) {
      buttons.loading('upload')
      let promises = []
      let progress = new Array(files.length).fill(0)

      let onupload = (id) => (event) => {
        progress[id] = (event.loaded / event.total) * 100

        let sum = 0
        for (let i = 0; i < progress.length; i++) {
          sum += progress[i]
        }

        this.$store.commit('setProgress', Math.ceil(sum / progress.length))
      }

      for (let i = 0; i < files.length; i++) {
        let file = files[i]
        promises.push(api.post(this.$route.path + base + file.name, file, overwrite, onupload(i)))
      }

      let finish = () => {
        buttons.success('upload')
        this.$store.commit('setProgress', 0)
      }

      Promise.all(promises)
        .then(() => {
          finish()
          this.$store.commit('setReload', true)
        })
        .catch(error => {
          finish()
          this.$showError(error)
        })

      return false
    },
    async sort (by) {
      let asc = false

      if (by === 'name') {
        if (this.nameIcon === 'arrow_upward') {
          asc = true
        }
      } else if (by === 'size') {
        if (this.sizeIcon === 'arrow_upward') {
          asc = true
        }
      } else if (by === 'modified') {
        if (this.modifiedIcon === 'arrow_upward') {
          asc = true
        }
      }

      try {
        await users.update({ id: this.user.id, sorting: { by, asc } }, ['sorting'])
      } catch (e) {
        this.$showError(e)
      }

      this.$store.commit('setReload', true)
    }
  }
}
</script>
