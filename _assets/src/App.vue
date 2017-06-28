<template>
  <div id="app">
    <header>
      <div id="top-bar">
            <img src="./assets/logo.svg" alt="File Manager">
            <search></search>
        </div>
        <div id="bottom-bar">
    
        </div>
        <div id="click-overlay"></div>
    </header>
    <nav id="sidebar">
      <a class="action" :href="baseURL()">
        <i class="material-icons">folder</i>
        <span>My Files</span>
      </a>
      <div class="action" id="logout" tabindex="0" role="button" aria-label="Log out">
        <i class="material-icons" title="Logout">exit_to_app</i>
        <span>Logout</span>
      </div>
    </nav>
    <main>
      <listing v-if="Kind == 'listing'"></listing> 
    </main>

    <preview v-if="Kind == 'preview'"></preview> 

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

export default {
  name: 'app',
  components: { Search, Preview, Listing },
  data: function () {
    return window.page
  },
  methods: {
    baseURL: function () {
      return window.data.baseURL + '/'
    }
  }
}
</script>

<style>
@import './css/styles.css';
</style>
