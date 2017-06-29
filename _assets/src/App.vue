<template>
  <div id="app" :class="{ multiple }">
    <header>
      <div>
        <img src="./assets/logo.svg" alt="File Manager">
        <search></search>
      </div>
      <div>
        <rename-button v-show="showRenameButton()"></rename-button>
        <move-button v-show="showMoveButton()"></move-button>
        <delete-button v-show="showDeleteButton()"></delete-button>
        <switch-button v-show="req.kind !== 'editor'"></switch-button>
        <download-button></download-button>
        <upload-button v-show="showUpload()"></upload-button>
        <info-button></info-button>
      </div>
      <!-- <div id="click-overlay"></div> -->
    </header>
    <nav id="sidebar">
      <a class="action" :href="baseURL + '/'">
        <i class="material-icons">folder</i>
        <span>My Files</span>
      </a>
      <div class="action" id="logout" tabindex="0" role="button" aria-label="Log out">
        <i class="material-icons" title="Logout">exit_to_app</i>
        <span>Logout</span>
      </div>
    </nav>
    <main>
      <listing v-if="req.kind === 'listing'"></listing> 
    </main>

    <preview v-if="req.kind === 'preview'"></preview> 
    
    <rename-prompt v-if="showRename" :class="{ active: showRename }"></rename-prompt>
    <delete-prompt v-if="showDelete" :class="{ active: showDelete }"></delete-prompt>
    <info-prompt v-if="showInfo" :class="{ active: showInfo }"></info-prompt>
    <move-prompt v-if="showMove" :class="{ active: showMove }"></move-prompt>
    <help v-show="showHelp" :class="{ active: showHelp }"></help>

    <div v-show="showOverlay()" @click="resetPrompts" class="overlay" :class="{ active: showOverlay() }"></div>

    <footer>Served with <a rel="noopener noreferrer" href="https://github.com/hacdias/caddy-filemanager">File Manager</a>.</footer>
  </div>
</template>

<script>
import Search from './components/Search'
import Preview from './components/Preview'
import Help from './components/Help'
import Listing from './components/Listing'
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
import css from './css.js'

var $ = window.info

function updateColumnSizes () {
  let columns = Math.floor(document.querySelector('main').offsetWidth / 300)
  let items = css(['#listing.mosaic .item', '.mosaic#listing .item'])

  items.style.width = `calc(${100 / columns}% - 1em)`
}

function resetPrompts () {
  $.showHelp = false
  $.showInfo = false
  $.showDelete = false
  $.showRename = false
  $.showMove = false
}

window.addEventListener('keydown', (event) => {
  // Esc!
  if (event.keyCode === 27) {
    resetPrompts()

    // Unselect all files and folders.
    if ($.req.kind === 'listing') {
      let items = document.getElementsByClassName('item')
      Array.from(items).forEach(link => {
        link.setAttribute('aria-selected', false)
      })

      $.selected = []
    }

    return
  }

  // Del!
  if (event.keyCode === 46) {
    $.showDelete = true
  }

  // F1!
  if (event.keyCode === 112) {
    event.preventDefault()
    $.showHelp = true
  }

  // F2!
  if (event.keyCode === 113) {
    $.showRename = true
  }

  // CTRL + S
  if (event.ctrlKey || event.metaKey) {
    switch (String.fromCharCode(event.which).toLowerCase()) {
      case 's':
        event.preventDefault()

        if ($.req.kind !== 'editor') {
          window.location = '?download=true'
          return
        }

        // TODO: save file on editor!
    }
  }
})

export default {
  name: 'app',
  components: {
    Search,
    Preview,
    Listing,
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
    MovePrompt
  },
  mounted: function () {
    updateColumnSizes()
    window.addEventListener('resize', updateColumnSizes)
    window.history.replaceState({
      url: window.location.pathname,
      name: document.title
    }, document.title, window.location.pathname)

    document.getElementById('loading').classList.add('done')
  },
  data: function () {
    return window.info
  },
  methods: {
    showOverlay: function () {
      return this.showInfo || this.showHelp || this.showDelete || this.showRename || this.showMove
    },
    showUpload: function () {
      if (this.req.kind === 'editor') return false
      return $.user.allowNew
    },
    showDeleteButton: function () {
      if (this.req.kind === 'listing') {
        if (this.selected.length === 0) {
          return false
        }

        return $.user.allowEdit
      }

      return $.user.allowEdit
    },
    showRenameButton: function () {
      if (this.req.kind === 'listing') {
        if (this.selected.length === 1) {
          return $.user.allowEdit
        }

        return false
      }

      return $.user.allowEdit
    },
    showMoveButton: function () {
      if (this.req.kind !== 'listing') {
        return false
      }

      if (this.selected.length > 0) {
        return $.user.allowEdit
      }

      return false
    },
    resetPrompts: resetPrompts
  }
}
</script>

<style>
@import './css/styles.css';
</style>
