<template>
  <div id="app">
    <header>
      <div id="first-bar">
        <img src="./assets/logo.svg" alt="File Manager">
        <search></search>
      </div>
      <div id="second-bar">
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
      <listing v-if="page.kind == 'listing'"></listing> 
    </main>

    <preview v-if="page.kind == 'preview'"></preview> 

    <div class="overlay"></div>
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

     <footer>Served with <a rel="noopener noreferrer" href="https://github.com/hacdias/caddy-filemanager">File Manager</a>.</footer>
  </div>
</template>

<script>
import Search from './components/Search'
import Preview from './components/Preview'
import Listing from './components/Listing'
import InfoButton from './components/InfoButton'
import css from './css.js'

function updateColumnSizes () {
  let columns = Math.floor(document.querySelector('main').offsetWidth / 300)
  let items = css(['#listing.mosaic .item', '.mosaic#listing .item'])

  items.style.width = `calc(${100 / columns}% - 1em)`
}

export default {
  name: 'app',
  components: { Search, Preview, Listing, InfoButton },
  mounted: function () {
    updateColumnSizes()
    window.addEventListener('resize', updateColumnSizes)
    window.history.replaceState({ url: window.location.pathname, name: document.title }, document.title, window.location.pathname)
  },
  data: function () {
    return window.info
  }
}
</script>

<style>
@import './css/styles.css';
</style>
