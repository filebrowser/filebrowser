<template>
  <div v-if="(req.numDirs + req.numFiles) == 0">
    <h2 class="message">
      <i class="material-icons">sentiment_dissatisfied</i>
      <span>{{ $t('files.lonely') }}</span>
    </h2>
    <input id="upload-input" style="display:none" type="file" multiple @change="uploadInput($event)">
  </div>
  <div
    v-else
    id="listing"
    :class="user.viewMode"
    @dragenter="dragEnter"
    @dragend="dragEnd"
  >
    <div>
      <div class="item header">
        <div />
        <div>
          <p
            :class="{ active: nameSorted }"
            class="name"
            role="button"
            tabindex="0"
            :title="$t('files.sortByName')"
            :aria-label="$t('files.sortByName')"
            @click="sort('name')"
          >
            <span>{{ $t('files.name') }}</span>
            <i class="material-icons">{{ nameIcon }}</i>
          </p>

          <p
            :class="{ active: sizeSorted }"
            class="size"
            role="button"
            tabindex="0"
            :title="$t('files.sortBySize')"
            :aria-label="$t('files.sortBySize')"
            @click="sort('size')"
          >
            <span>{{ $t('files.size') }}</span>
            <i class="material-icons">{{ sizeIcon }}</i>
          </p>
          <p
            :class="{ active: modifiedSorted }"
            class="modified"
            role="button"
            tabindex="0"
            :title="$t('files.sortByLastModified')"
            :aria-label="$t('files.sortByLastModified')"
            @click="sort('modified')"
          >
            <span>{{ $t('files.lastModified') }}</span>
            <i class="material-icons">{{ modifiedIcon }}</i>
          </p>
        </div>
      </div>
    </div>

    <h2 v-if="req.numDirs > 0">{{ $t('files.folders') }}</h2>
    <div v-if="req.numDirs > 0">
      <item
        v-for="(item) in dirs"
        :key="base64(item.name)"
        :index="item.index"
        :name="item.name"
        :is-dir="item.isDir"
        :url="item.url"
        :modified="item.modified"
        :type="item.type"
        :size="item.size"
      />
    </div>

    <h2 v-if="req.numFiles > 0">{{ $t('files.files') }}</h2>
    <div v-if="req.numFiles > 0">
      <item
        v-for="(item) in files"
        :key="base64(item.name)"
        :index="item.index"
        :name="item.name"
        :is-dir="item.isDir"
        :url="item.url"
        :modified="item.modified"
        :type="item.type"
        :size="item.size"
      />
    </div>

    <input id="upload-input" style="display:none" type="file" multiple @change="uploadInput($event)">

    <div id="multiple-selection" :class="{ active: $store.state.multiple }">
      <p>{{ $t('files.multipleSelectionEnabled') }}</p>
      <div tabindex="0" role="button" :title="$t('files.clear')" :aria-label="$t('files.clear')" class="action" @click="$store.commit('multiple', false)">
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
import url from '@/utils/url'

export default {
  name: 'Listing',
  components: { Item },
  data: function() {
    return {
      show: 50
    }
  },
  computed: {
    ...mapState(['req', 'selected', 'user']),
    nameSorted() {
      return (this.req.sorting.by === 'name')
    },
    sizeSorted() {
      return (this.req.sorting.by === 'size')
    },
    modifiedSorted() {
      return (this.req.sorting.by === 'modified')
    },
    ascOrdered() {
      return this.req.sorting.asc
    },
    items() {
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
    dirs() {
      return this.items.dirs.slice(0, this.show)
    },
    files() {
      let show = this.show - this.items.dirs.length

      if (show < 0) show = 0

      return this.items.files.slice(0, show)
    },
    nameIcon() {
      if (this.nameSorted && !this.ascOrdered) {
        return 'arrow_upward'
      }

      return 'arrow_downward'
    },
    sizeIcon() {
      if (this.sizeSorted && this.ascOrdered) {
        return 'arrow_downward'
      }

      return 'arrow_upward'
    },
    modifiedIcon() {
      if (this.modifiedSorted && this.ascOrdered) {
        return 'arrow_downward'
      }

      return 'arrow_upward'
    }
  },
  mounted: function() {
    // Check the columns size for the first time.
    this.resizeEvent()

    // Add the needed event listeners to the window and document.
    window.addEventListener('keydown', this.keyEvent)
    window.addEventListener('resize', this.resizeEvent)
    window.addEventListener('scroll', this.scrollEvent)
    document.addEventListener('dragover', this.preventDefault)
    document.addEventListener('drop', this.drop)
  },
  beforeDestroy() {
    // Remove event listeners before destroying this page.
    window.removeEventListener('keydown', this.keyEvent)
    window.removeEventListener('resize', this.resizeEvent)
    window.removeEventListener('scroll', this.scrollEvent)
    document.removeEventListener('dragover', this.preventDefault)
    document.removeEventListener('drop', this.drop)
  },
  methods: {
    ...mapMutations(['updateUser']),
    base64: function(name) {
      return window.btoa(unescape(encodeURIComponent(name)))
    },
    keyEvent(event) {
      if (!event.ctrlKey && !event.metaKey) {
        return
      }

      const key = String.fromCharCode(event.which).toLowerCase()

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
    preventDefault(event) {
      // Wrapper around prevent default.
      event.preventDefault()
    },
    copyCut(event, key) {
      if (event.target.tagName.toLowerCase() === 'input') {
        return
      }

      const items = []

      for (const i of this.selected) {
        items.push({
          from: this.req.items[i].url,
          name: encodeURIComponent(this.req.items[i].name)
        })
      }

      if (items.length === 0) {
        return
      }

      this.$store.commit('updateClipboard', {
        key: key,
        items: items
      })
    },
    paste(event) {
      if (event.target.tagName.toLowerCase() === 'input') {
        return
      }

      const items = []

      for (const item of this.$store.state.clipboard.items) {
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
    resizeEvent() {
      // Update the columns size based on the window width.
      let columns = Math.floor(document.querySelector('main').offsetWidth / 300)
      const items = css(['#listing.mosaic .item', '.mosaic#listing .item'])
      if (columns === 0) columns = 1
      items.style.width = `calc(${100 / columns}% - 1em)`
    },
    scrollEvent() {
      if ((window.innerHeight + window.scrollY) >= document.body.offsetHeight) {
        this.show += 50
      }
    },
    dragEnter() {
      // When the user starts dragging an item, put every
      // file on the listing with 50% opacity.
      const items = document.getElementsByClassName('item')

      Array.from(items).forEach(file => {
        file.style.opacity = 0.5
      })
    },
    dragEnd() {
      this.resetOpacity()
    },
    drop: function(event) {
      event.preventDefault()
      this.resetOpacity()

      const dt = event.dataTransfer
      const files = dt.files
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
    checkConflict(files, items, base) {
      if (typeof items === 'undefined' || items === null) {
        items = []
      }

      let conflict = false
      for (let i = 0; i < files.length; i++) {
        const res = items.findIndex(function hasConflict(element) {
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
    uploadInput(event) {
      this.checkConflict(event.currentTarget.files, this.req.items, '')
    },
    resetOpacity() {
      const items = document.getElementsByClassName('item')

      Array.from(items).forEach(file => {
        file.style.opacity = 1
      })
    },
    handleFiles(files, base, overwrite = false) {
      buttons.loading('upload')
      const promises = []
      const progress = new Array(files.length).fill(0)

      const onupload = (id) => (event) => {
        progress[id] = (event.loaded / event.total) * 100

        let sum = 0
        for (let i = 0; i < progress.length; i++) {
          sum += progress[i]
        }

        this.$store.commit('setProgress', Math.ceil(sum / progress.length))
      }

      for (let i = 0; i < files.length; i++) {
        const file = files[i]
        const filenameEncoded = url.encodeRFC5987ValueChars(file.name)
        promises.push(api.post(this.$route.path + base + filenameEncoded, file, overwrite, onupload(i)))
      }

      const finish = () => {
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
    async sort(by) {
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
        await users.update({ id: this.user.id, sorting: { by, asc }}, ['sorting'])
      } catch (e) {
        this.$showError(e)
      }

      this.$store.commit('setReload', true)
    }
  }
}
</script>
