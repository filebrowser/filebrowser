<template>
  <div :class="{ multiple }">
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

    <nav>
      <router-link class="action" to="/files/">
        <i class="material-icons">folder</i>
        <span>My Files</span>
      </router-link>

      <div v-if="user.allowNew">
        <button @click="$store.commit('showNewDir', true)" aria-label="New directory" title="New directory" class="action">
          <i class="material-icons">create_new_folder</i>
          <span>New folder</span>
        </button>
        <button @click="$store.commit('showNewFile', true)" aria-label="New file" title="New file" class="action">
          <i class="material-icons">note_add</i>
          <span>New file</span>
        </button>
      </div>

      <div v-for="plugin in plugins">
        <button v-for="action in plugin.sidebar" @click="action.click" :aria-label="action.name" :title="action.name" class="action">
          <i class="material-icons">{{ action.icon }}</i>
          <span>{{ action.name }}</span>
        </button>
      </div>

      <button @click="logout" class="action" id="logout" aria-label="Log out">
        <i class="material-icons" title="Logout">exit_to_app</i>
        <span>Logout</span>
      </button>
    </nav>

    <main>
      <editor v-if="isEditor"></editor>
      <listing v-if="isListing"></listing>
      <preview v-if="isPreview"></preview>
    </main>

    <download-prompt v-if="showDownload" :class="{ active: showDownload }"></download-prompt>
    <new-file-prompt v-if="showNewFile" :class="{ active: showNewFile }"></new-file-prompt>
    <new-dir-prompt v-if="showNewDir" :class="{ active: showNewDir }"></new-dir-prompt>
    <rename-prompt v-if="showRename" :class="{ active: showRename }"></rename-prompt>
    <delete-prompt v-if="showDelete" :class="{ active: showDelete }"></delete-prompt>
    <info-prompt v-if="showInfo" :class="{ active: showInfo }"></info-prompt>
    <move-prompt v-if="showMove" :class="{ active: showMove }"></move-prompt>
    <help v-show="showHelp" :class="{ active: showHelp }"></help>
    <div v-show="showOverlay" @click="resetPrompts" class="overlay" :class="{ active: showOverlay }"></div>

    <footer>Served with <a rel="noopener noreferrer" href="https://github.com/hacdias/caddy-filemanager">File Manager</a>.</footer>
  </div>
</template>

<script>
import Search from './Search'
import Help from './Help'
import Preview from './Preview'
import Listing from './Listing'
import Editor from './Editor'
import InfoButton from './buttons/InfoButton'
import InfoPrompt from './prompts/InfoPrompt'
import DeleteButton from './buttons/DeleteButton'
import DeletePrompt from './prompts/DeletePrompt'
import RenameButton from './buttons/RenameButton'
import RenamePrompt from './prompts/RenamePrompt'
import UploadButton from './buttons/UploadButton'
import DownloadButton from './buttons/DownloadButton'
import DownloadPrompt from './prompts/DownloadPrompt'
import SwitchButton from './buttons/SwitchViewButton'
import MoveButton from './buttons/MoveButton'
import MovePrompt from './prompts/MovePrompt'
import NewFilePrompt from './prompts/NewFilePrompt'
import NewDirPrompt from './prompts/NewDirPrompt'
import css from '@/utils/css'
import auth from '@/utils/auth'
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
    InfoButton,
    InfoPrompt,
    Help,
    DeleteButton,
    DeletePrompt,
    RenameButton,
    RenamePrompt,
    DownloadButton,
    DownloadPrompt,
    UploadButton,
    SwitchButton,
    MoveButton,
    MovePrompt,
    NewFilePrompt,
    NewDirPrompt
  },
  computed: {
    ...mapGetters([
      'selectedCount',
      'showOverlay'
    ]),
    ...mapState([
      'req',
      'user',
      'baseURL',
      'multiple',
      'showInfo',
      'showHelp',
      'showDelete',
      'showRename',
      'showMove',
      'showNewFile',
      'showNewDir',
      'showDownload'
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
      plugins: [],
      loading: true
    }
  },
  created () {
    this.fetchData()
  },
  watch: {
    '$route': 'fetchData'
  },
  mounted () {
    updateColumnSizes()
    window.addEventListener('resize', updateColumnSizes)

    if (window.plugins !== undefined || window.plugins !== null) {
      this.plugins = window.plugins
    }

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
    },
    resetPrompts: function () {
      this.$store.commit('resetPrompts')
    },
    logout: auth.logout
  }
}
</script>
