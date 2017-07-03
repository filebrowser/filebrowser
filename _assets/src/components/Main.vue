<template>
  <div :class="{ multiple, loading }">
    <header>
      <div>
        <img src="../assets/logo.svg" alt="File Manager">
        <search></search>
      </div>
      <div>
        <rename-button v-show="!loading && showRenameButton()"></rename-button>
        <move-button v-show="!loading && showMoveButton()"></move-button>
        <delete-button v-show="!loading && showDeleteButton()"></delete-button>
        <switch-button v-show="!loading && req.kind !== 'editor'"></switch-button>
        <download-button></download-button>
        <upload-button v-show="!loading && showUpload()"></upload-button>
        <info-button></info-button>

        <button v-show="isListing" @click="$store.commit('multiple', true)" aria-label="Select multiple" class="action">
          <i class="material-icons">check_circle</i>
          <span>Select</span>
        </button>
      </div>
    </header>

    <sidebar></sidebar>

    <main>
      <div v-if="loading">Loading...</div>
      <div v-else-if="error">
        <h2 class="message" v-if="error === 404">
          <i class="material-icons">gps_off</i>
          <span>This location can't be reached.</span>
        </h2>
        <h2 class="message" v-else-if="error === 403">
          <i class="material-icons">error</i>
          <span>You're not welcome here.</span>
        </h2>
      </div>
      <editor v-else-if="isEditor"></editor>
      <listing v-else-if="isListing"></listing>
      <preview v-else-if="isPreview"></preview>
    </main>

    <prompts></prompts>
  </div>
</template>

<script>
import Search from './Search'
import Preview from './Preview'
import Listing from './Listing'
import Editor from './Editor'
import Sidebar from './Sidebar'
import Prompts from './prompts/Prompts'
import InfoButton from './buttons/InfoButton'
import DeleteButton from './buttons/DeleteButton'
import RenameButton from './buttons/RenameButton'
import UploadButton from './buttons/UploadButton'
import DownloadButton from './buttons/DownloadButton'
import SwitchButton from './buttons/SwitchViewButton'
import MoveButton from './buttons/MoveButton'
import css from '@/utils/css'
import api from '@/utils/api'
import {mapGetters, mapState} from 'vuex'

function updateColumnSizes () {
  let columns = Math.floor(document.querySelector('main').offsetWidth / 300)
  let items = css(['#listing.mosaic .item', '.mosaic#listing .item'])

  if (columns === 0) columns = 1

  items.style.width = `calc(${100 / columns}% - 1em)`
}

export default {
  name: 'main',
  components: {
    Search,
    Preview,
    Listing,
    Editor,
    Sidebar,
    InfoButton,
    DeleteButton,
    RenameButton,
    DownloadButton,
    UploadButton,
    SwitchButton,
    MoveButton,
    Prompts
  },
  computed: {
    ...mapGetters([
      'selectedCount'
    ]),
    ...mapState([
      'req',
      'user',
      'reload',
      'baseURL',
      'multiple'
    ]),
    isListing () {
      return this.req.kind === 'listing' && !this.loading
    },
    isPreview () {
      return this.req.kind === 'preview' && !this.loading
    },
    isEditor () {
      return this.req.kind === 'editor' && !this.loading
    }
  },
  data: function () {
    return {
      loading: true,
      error: null
    }
  },
  created () {
    this.fetchData()
  },
  watch: {
    '$route': 'fetchData',
    'reload': function () {
      this.$store.commit('setReload', false)
      this.fetchData()
    }
  },
  mounted () {
    updateColumnSizes()
    window.addEventListener('resize', updateColumnSizes)

    window.addEventListener('keydown', (event) => {
      // Esc!
      if (event.keyCode === 27) {
        this.$store.commit('resetPrompts')

        // Unselect all files and folders.
        if (this.req.kind === 'listing') {
          let items = document.getElementsByClassName('item')
          Array.from(items).forEach(link => {
            link.setAttribute('aria-selected', false)
          })

          this.$store.commit('resetSelected')
        }

        return
      }

      // Del!
      if (event.keyCode === 46) {
        if (this.showDeleteButton()) {
          this.$store.commit('showDelete', true)
        }
      }

      // F1!
      if (event.keyCode === 112) {
        event.preventDefault()
        this.$store.commit('showHelp', true)
      }

      // F2!
      if (event.keyCode === 113) {
        if (this.showRenameButton()) {
          this.$store.commit('showRename', true)
        }
      }

      // CTRL + S
      if (event.ctrlKey || event.metaKey) {
        switch (String.fromCharCode(event.which).toLowerCase()) {
          case 's':
            event.preventDefault()

            if (this.req.kind !== 'editor') {
              window.location = '?download=true'
              return
            }

            // TODO: save file on editor!
        }
      }
    })
  },
  methods: {
    fetchData () {
      this.loading = true
      this.error = null
      // Reset selected items and multiple selection.
      this.$store.commit('resetSelected')
      this.$store.commit('multiple', false)

      let url = this.$route.path
      if (url === '') url = '/'
      if (url[0] !== '/') url = '/' + url

      console.log('Going to ' + url)

      api.fetch(url)
      .then((trueURL) => {
        if (!url.endsWith('/') && trueURL.endsWith('/')) {
          window.history.replaceState(window.history.state, document.title, window.location.pathname + '/')
        }

        this.loading = false
      })
      .catch(error => {
        // TODO: 404, 403 and 500!
        console.log(error)
        this.error = error
        this.loading = false
      })
    },
    showUpload: function () {
      if (this.req.kind === 'editor') return false
      return this.user.allowNew
    },
    showDeleteButton: function () {
      if (this.req.kind === 'listing') {
        if (this.selectedCount === 0) {
          return false
        }

        return this.user.allowEdit
      }

      return this.user.allowEdit
    },
    showRenameButton: function () {
      if (this.req.kind === 'listing') {
        if (this.selectedCount === 1) {
          return this.user.allowEdit
        }

        return false
      }

      return this.user.allowEdit
    },
    showMoveButton: function () {
      if (this.req.kind !== 'listing') {
        return false
      }

      if (this.selectedCount > 0) {
        return this.user.allowEdit
      }

      return false
    }
  }
}
</script>
