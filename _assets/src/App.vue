<template>
  <div id="app" :class="{ multiple: $store.state.multiple }">
    <header>
      <div>
        <img src="./assets/logo.svg" alt="File Manager">
        <search></search>
      </div>
      <div>
        <rename-button v-show="showRenameButton()"></rename-button>
        <move-button v-show="showMoveButton()"></move-button>
        <delete-button v-show="showDeleteButton()"></delete-button>
        <switch-button v-show="$store.state.req.kind !== 'editor'"></switch-button>
        <download-button></download-button>
        <upload-button v-show="showUpload()"></upload-button>
        <info-button></info-button>
      </div>
    </header>

    <nav>
      <a class="action" :href="$store.state.baseURL + '/'">
        <i class="material-icons">folder</i>
        <span>My Files</span>
      </a>
      <div v-if="$store.state.user.allowNew">
        <button @click="$store.commit('showNewDir', true)" aria-label="New directory" title="New directory" class="action">
          <i class="material-icons">create_new_folder</i>
          <span>New folder</span>
        </button>
        <button @click="$store.commit('showNewFile', true)" aria-label="New file" title="New file" class="action">
          <i class="material-icons">note_add</i>
          <span>New file</span>
        </button>
      </div>
      <button class="action" id="logout" tabindex="0" role="button" aria-label="Log out">
        <i class="material-icons" title="Logout">exit_to_app</i>
        <span>Logout</span>
      </button>
    </nav>

    <main>
      <editor v-if="$store.state.req.kind === 'editor'"></editor>
      <listing v-if="$store.state.req.kind === 'listing'"></listing> 
      <preview v-if="$store.state.req.kind === 'preview'"></preview> 
    </main>
    
    <new-file-prompt v-if="$store.state.showNewFile" :class="{ active: $store.state.showNewFile }"></new-file-prompt>
    <new-dir-prompt v-if="$store.state.showNewDir" :class="{ active: $store.state.showNewDir }"></new-dir-prompt>
    <rename-prompt v-if="$store.state.showRename" :class="{ active: $store.state.showRename }"></rename-prompt>
    <delete-prompt v-if="$store.state.showDelete" :class="{ active: $store.state.showDelete }"></delete-prompt>
    <info-prompt v-if="$store.state.showInfo" :class="{ active: $store.state.showInfo }"></info-prompt>
    <move-prompt v-if="$store.state.showMove" :class="{ active: $store.state.showMove }"></move-prompt>
    <help v-show="$store.state.showHelp" :class="{ active: $store.state.showHelp }"></help>
    <div v-show="$store.getters.showOverlay" @click="resetPrompts" class="overlay" :class="{ active: $store.getters.showOverlay }"></div>

    <footer>Served with <a rel="noopener noreferrer" href="https://github.com/hacdias/caddy-filemanager">File Manager</a>.</footer>
  </div>
</template>

<script>
import Search from './components/Search'
import Help from './components/Help'
import Preview from './components/Preview'
import Listing from './components/Listing'
import Editor from './components/Editor'
import InfoButton from './components/InfoButton'
import InfoPrompt from './components/InfoPrompt'
import DeleteButton from './components/DeleteButton'
import DeletePrompt from './components/DeletePrompt'
import RenameButton from './components/RenameButton'
import RenamePrompt from './components/RenamePrompt'
import UploadButton from './components/UploadButton'
import DownloadButton from './components/DownloadButton'
import SwitchButton from './components/SwitchViewButton'
import MoveButton from './components/MoveButton'
import MovePrompt from './components/MovePrompt'
import NewFilePrompt from './components/NewFilePrompt'
import NewDirPrompt from './components/NewDirPrompt'
import css from './css.js'

function updateColumnSizes () {
  let columns = Math.floor(document.querySelector('main').offsetWidth / 300)
  let items = css(['#listing.mosaic .item', '.mosaic#listing .item'])

  if (columns === 0) columns = 1

  items.style.width = `calc(${100 / columns}% - 1em)`
}

export default {
  name: 'app',
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
    UploadButton,
    SwitchButton,
    MoveButton,
    MovePrompt,
    NewFilePrompt,
    NewDirPrompt
  },
  mounted: function () {
    updateColumnSizes()
    window.addEventListener('resize', updateColumnSizes)
    window.history.replaceState({
      url: window.location.pathname,
      name: document.title
    }, document.title, window.location.pathname)

    window.addEventListener('popstate', (event) => {
      event.preventDefault()
      event.stopPropagation()

      this.$store.commit('multiple', false)
      this.$store.commit('resetSelected')
      this.$store.commit('resetPrompts')

      let request = new window.XMLHttpRequest()
      request.open('GET', event.state.url, true)
      request.setRequestHeader('Accept', 'application/json')

      request.onload = () => {
        if (request.status === 200) {
          let req = JSON.parse(request.responseText)
          this.$store.commit('updateRequest', req)
          document.title = event.state.name
        } else {
          console.log(request.responseText)
        }
      }

      request.onerror = (error) => { console.log(error) }
      request.send()
    })

    window.addEventListener('keydown', (event) => {
      // Esc!
      if (event.keyCode === 27) {
        this.$store.commit('resetPrompts')

        // Unselect all files and folders.
        if (this.$store.state.req.kind === 'listing') {
          let items = document.getElementsByClassName('item')
          Array.from(items).forEach(link => {
            link.setAttribute('aria-selected', false)
          })

          this.$store.selected = []
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

            if (this.$store.state.req.kind !== 'editor') {
              window.location = '?download=true'
              return
            }

            // TODO: save file on editor!
        }
      }
    })

    let loading = document.getElementById('loading')
    loading.classList.add('done')

    setTimeout(function () {
      loading.parentNode.removeChild(loading)
    }, 1000)
  },
  methods: {
    showUpload: function () {
      if (this.$store.state.req.kind === 'editor') return false
      return this.$store.state.user.allowNew
    },
    showDeleteButton: function () {
      if (this.$store.state.req.kind === 'listing') {
        if (this.$store.getters.selectedCount === 0) {
          return false
        }

        return this.$store.state.user.allowEdit
      }

      return this.$store.state.user.allowEdit
    },
    showRenameButton: function () {
      if (this.$store.state.req.kind === 'listing') {
        if (this.$store.getters.selectedCount === 1) {
          return this.$store.state.user.allowEdit
        }

        return false
      }

      return this.$store.state.user.allowEdit
    },
    showMoveButton: function () {
      if (this.$store.state.req.kind !== 'listing') {
        return false
      }

      if (this.$store.getters.selectedCount > 0) {
        return this.$store.state.user.allowEdit
      }

      return false
    },
    resetPrompts: function () {
      this.$store.commit('resetPrompts')
    }
  }
}
</script>

<style>
@import './css/styles.css';
</style>
