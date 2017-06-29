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
      <div v-if="user.allowNew">
        <button @click="showNewDir = true" aria-label="New directory" title="New directory" class="action">
          <i class="material-icons">create_new_folder</i>
          <span>New folder</span>
        </button>
        <button @click="showNewFile = true" aria-label="New file" title="New file" class="action">
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
      <listing v-if="req.kind === 'listing'"></listing> 
      <preview v-if="req.kind === 'preview'"></preview> 
    </main>
    
    <new-file-prompt v-if="showNewFile" :class="{ active: showNewFile }"></new-file-prompt>
    <new-dir-prompt v-if="showNewDir" :class="{ active: showNewDir }"></new-dir-prompt>
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
import NewFilePrompt from './components/NewFilePrompt'
import NewDirPrompt from './components/NewDirPrompt'
import css from './css.js'

var $ = window.info

function updateColumnSizes () {
  let columns = Math.floor(document.querySelector('main').offsetWidth / 300)
  let items = css(['#listing.mosaic .item', '.mosaic#listing .item'])

  if (columns === 0) columns = 1

  items.style.width = `calc(${100 / columns}% - 1em)`
}

function resetPrompts () {
  $.showHelp = false
  $.showInfo = false
  $.showDelete = false
  $.showRename = false
  $.showMove = false
  $.showNewFile = false
  $.showNewDir = false
}

function showRenameButton () {
  if ($.req.kind === 'listing') {
    if ($.selected.length === 1) {
      return $.user.allowEdit
    }

    return false
  }

  return $.user.allowEdit
}

function showDeleteButton () {
  if ($.req.kind === 'listing') {
    if ($.selected.length === 0) {
      return false
    }

    return $.user.allowEdit
  }

  return $.user.allowEdit
}

function keydown (event) {
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
    if (showDeleteButton()) {
      $.showDelete = true
    }
  }

  // F1!
  if (event.keyCode === 112) {
    event.preventDefault()
    $.showHelp = true
  }

  // F2!
  if (event.keyCode === 113) {
    if (showRenameButton()) {
      $.showRename = true
    }
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
}

function startup () {
  updateColumnSizes()
  window.addEventListener('resize', updateColumnSizes)
  window.history.replaceState({
    url: window.location.pathname,
    name: document.title
  }, document.title, window.location.pathname)

  window.addEventListener('keydown', keydown)

  let loading = document.getElementById('loading')
  loading.classList.add('done')

  setTimeout(function () {
    loading.parentNode.removeChild(loading)
  }, 1000)
}

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
    MovePrompt,
    NewFilePrompt,
    NewDirPrompt
  },
  mounted: function () {
    startup()
  },
  data: function () {
    return window.info
  },
  methods: {
    showOverlay: function () {
      return $.showInfo ||
        $.showHelp ||
        $.showDelete ||
        $.showRename ||
        $.showMove ||
        $.showNewFile ||
        $.showNewDir
    },
    showUpload: function () {
      if (this.req.kind === 'editor') return false
      return $.user.allowNew
    },
    showDeleteButton: showDeleteButton,
    showRenameButton: showRenameButton,
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
