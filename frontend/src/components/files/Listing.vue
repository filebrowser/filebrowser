<template>
  <div v-if="(req.numDirs + req.numFiles) == 0">
    <h2 class="message">
      <i class="material-icons">sentiment_dissatisfied</i>
      <span>{{ $t('files.lonely') }}</span>
    </h2>
    <input style="display:none" type="file" id="upload-input" @change="uploadInput($event)" multiple>
    <input style="display:none" type="file" id="upload-folder-input" @change="uploadInput($event)" webkitdirectory multiple>
  </div>
  <div v-else id="listing"
    :class="user.viewMode">
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
    <input style="display:none" type="file" id="upload-folder-input" @change="uploadInput($event)" webkitdirectory multiple>

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
import * as upload  from '@/utils/upload'

export default {
  name: 'listing',
  components: { Item },
  data: function () {
    return {
      showLimit: 50,
      dragCounter: 0
    }
  },
  computed: {
    ...mapState(['req', 'selected', 'user', 'show']),
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
      return this.items.dirs.slice(0, this.showLimit)
    },
    files () {
      let showLimit = this.showLimit - this.items.dirs.length

      if (showLimit < 0) showLimit = 0

      return this.items.files.slice(0, showLimit)
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
    document.addEventListener('dragenter', this.dragEnter)
    document.addEventListener('dragleave', this.dragLeave)
    document.addEventListener('drop', this.drop)
  },
  beforeDestroy () {
    // Remove event listeners before destroying this page.
    window.removeEventListener('keydown', this.keyEvent)
    window.removeEventListener('resize', this.resizeEvent)
    window.removeEventListener('scroll', this.scrollEvent)
    document.removeEventListener('dragover', this.preventDefault)
    document.removeEventListener('dragenter', this.dragEnter)
    document.removeEventListener('dragleave', this.dragLeave)
    document.removeEventListener('drop', this.drop)
  },
  methods: {
    ...mapMutations([ 'updateUser', 'addSelected' ]),
    base64: function (name) {
      return window.btoa(unescape(encodeURIComponent(name)))
    },
    keyEvent (event) {
      if (this.show !== null) {
        return
      }

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
        case 'a':
          event.preventDefault()
          for (let file of this.items.files) {
            if (this.$store.state.selected.indexOf(file.index) === -1) {
              this.addSelected(file.index)
            }
          }
          for (let dir of this.items.dirs) {
            if (this.$store.state.selected.indexOf(dir.index) === -1) {
              this.addSelected(dir.index)
            }
          }
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
        items: items,
        path: this.$route.path
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
        items.push({ from, to, name: item.name })
      }

      if (items.length === 0) {
        return
      }

      let action = (overwrite, rename) => {
        api.copy(items, overwrite, rename).then(() => {
          this.$store.commit('setReload', true)
        }).catch(this.$showError)
      }

      if (this.$store.state.clipboard.key === 'x') {
        action = (overwrite, rename) => {
          api.move(items, overwrite, rename).then(() => {
            this.$store.commit('resetClipboard')
            this.$store.commit('setReload', true)
          }).catch(this.$showError)
        }
      }

      if (this.$store.state.clipboard.path == this.$route.path) {
        action(false, true)

        return
      }

      let conflict = upload.checkConflict(items, this.req.items)

      let overwrite = false
      let rename = false

      if (conflict) {
        this.$store.commit('showHover', {
          prompt: 'replace-rename',
          confirm: (event, option) => {
            overwrite = option == 'overwrite'
            rename = option == 'rename'

            event.preventDefault()
            this.$store.commit('closeHovers')
            action(overwrite, rename)
          }
        })

        return
      }

      action(overwrite, rename)
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
        this.showLimit += 50
      }
    },
    dragEnter () {
      this.dragCounter++

      // When the user starts dragging an item, put every
      // file on the listing with 50% opacity.
      let items = document.getElementsByClassName('item')

      Array.from(items).forEach(file => {
        file.style.opacity = 0.5
      })
    },
    dragLeave () {
      this.dragCounter--

      if (this.dragCounter == 0) {
        this.resetOpacity()
      }
    },
    drop: async function (event) {
      event.preventDefault()
      this.dragCounter = 0
      this.resetOpacity()

      let dt = event.dataTransfer
      let el = event.target

      if (dt.files.length <= 0) return

      for (let i = 0; i < 5; i++) {
        if (el !== null && !el.classList.contains('item')) {
          el = el.parentElement
        }
      }

      let base = ''
      if (el !== null && el.classList.contains('item') && el.dataset.dir === 'true') {
        base = el.querySelector('.name').innerHTML + '/'
      }

      let files = await upload.scanFiles(dt)
      let path = this.$route.path.endsWith('/') ? this.$route.path + base : this.$route.path + '/' + base
      let items = this.req.items

      if (base !== '') {
        try {
          items = (await api.fetch(path)).items
        } catch (error) {
          this.$showError(error)
        }
      }

      let conflict = upload.checkConflict(files, items)

      if (conflict) {
        this.$store.commit('showHover', {
          prompt: 'replace',
          confirm: (event) => {
            event.preventDefault()
            this.$store.commit('closeHovers')
            upload.handleFiles(files, path, true)
          }
        })

        return
      }

      upload.handleFiles(files, path)
    },
    uploadInput (event) {
      this.$store.commit('closeHovers')

      let files = event.currentTarget.files
      let folder_upload = files[0].webkitRelativePath !== undefined && files[0].webkitRelativePath !== ''

      if (folder_upload) {
        for (let i = 0; i < files.length; i++) {
          let file = files[i]
          files[i].fullPath = file.webkitRelativePath
        }
      }

      let path = this.$route.path.endsWith('/') ? this.$route.path : this.$route.path + '/'
      let conflict = upload.checkConflict(files, this.req.items)

      if (conflict) {
        this.$store.commit('showHover', {
          prompt: 'replace',
          confirm: (event) => {
            event.preventDefault()
            this.$store.commit('closeHovers')
            upload.handleFiles(files, path, true)
          }
        })

        return
      }

      upload.handleFiles(files, path)
    },
    resetOpacity () {
      let items = document.getElementsByClassName('item')

      Array.from(items).forEach(file => {
        file.style.opacity = 1
      })
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
