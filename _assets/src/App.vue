<template>
  <div id="app">
    <header>
      <div>
        <img src="./assets/logo.svg" alt="File Manager">
        <search></search>
      </div>
      <div>
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
      <listing v-if="req.kind == 'listing'"></listing> 
    </main>

    <preview v-if="req.kind == 'preview'"></preview> 

    
    <!-- TODO: show on listing and allowedit -->
    <div class="floating">
        <div tabindex="0" role="button" class="action" id="new">
            <i class="material-icons" title="New file or directory">add</i>
        </div>
    </div>

    <!-- TODO ??? -->
     <div id="multiple-selection" class="mobile-only">
        <p>Multiple selection enabled</p>
        <div tabindex="0" role="button" class="action" id="multiple-selection-cancel">
            <i class="material-icons" title="Clear">clear</i>
        </div>
    </div>

    <info-prompt v-show="showInfo" :class="{ active: showInfo }"></info-prompt>
    <help v-show="showHelp" :class="{ active: showHelp }"></help>

    <div v-show="showOverlay()" class="overlay" :class="{ active: showOverlay() }"></div>

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
import css from './css.js'

function updateColumnSizes () {
  let columns = Math.floor(document.querySelector('main').offsetWidth / 300)
  let items = css(['#listing.mosaic .item', '.mosaic#listing .item'])

  items.style.width = `calc(${100 / columns}% - 1em)`
}

window.addEventListener('keydown', (event) => {
  // Esc!
  if (event.keyCode === 27) {
    window.info.showHelp = false
    window.info.showInfo = false
    window.info.showDelete = false
    window.info.showRename = false

    // Unselect all files and folders.
    if (window.info.req.kind === 'listing') {
      let items = document.getElementsByClassName('item')
      Array.from(items).forEach(link => {
        link.setAttribute('aria-selected', false)
      })

      window.info.listing.selected.length = 0
    }

    return
  }

  // Del!
  if (event.keyCode === 46) {
    window.info.showDelete = true
  }

  // F1!
  if (event.keyCode === 112) {
    event.preventDefault()
    window.info.showHelp = true
  }

  // F2!
  if (event.keyCode === 113) {
    window.info.showRename = true
  }

  // CTRL + S
  if (event.ctrlKey || event.metaKey) {
    switch (String.fromCharCode(event.which).toLowerCase()) {
      case 's':
        event.preventDefault()

        if (window.info.req.kind !== 'editor') {
          window.location = '?download=true'
          return
        }

        // TODO: save file on editor!
    }
  }
})

export default {
  name: 'app',
  components: { Search, Preview, Listing, InfoButton, InfoPrompt, Help },
  mounted: function () {
    updateColumnSizes()
    window.addEventListener('resize', updateColumnSizes)
    window.history.replaceState({ url: window.location.pathname, name: document.title }, document.title, window.location.pathname)
  },
  data: function () {
    return window.info
  },
  methods: {
    showOverlay: function () {
      return this.showInfo || this.showHelp
    }
  }
}
</script>

<style>
@import './css/styles.css';
</style>
