<template>
  <nav :class="{active}">
    <router-link class="action" to="/files/" aria-label="My Files" title="My Files">
      <i class="material-icons">folder</i>
      <span>My Files</span>
    </router-link>

    <div v-if="user.allowNew">
      <button @click="$store.commit('showHover', 'newDir')" aria-label="New directory" title="New directory" class="action">
        <i class="material-icons">create_new_folder</i>
        <span>New folder</span>
      </button>

      <button @click="$store.commit('showHover', 'newFile')" aria-label="New file" title="New file" class="action">
        <i class="material-icons">note_add</i>
        <span>New file</span>
      </button>
    </div>

    <div v-for="plugin in plugins" :key="plugin.name">
      <button v-for="action in plugin.sidebar" @click="action.click($event, pluginData, $route)" :aria-label="action.name" :title="action.name" :key="action.name" class="action">
        <i class="material-icons">{{ action.icon }}</i>
        <span>{{ action.name }}</span>
      </button>
    </div>

    <div>
      <router-link class="action" to="/settings" aria-label="Settings" title="Settings">
        <i class="material-icons">settings_applications</i>
        <span>Settings</span>
      </router-link>

      <button @click="logout" class="action" id="logout" aria-label="Log out" title="Logout">
        <i class="material-icons">exit_to_app</i>
        <span>Logout</span>
      </button>
    </div>

    <p class="credits">
      <span>Served with <a rel="noopener noreferrer" href="https://github.com/hacdias/caddy-filemanager">File Manager</a>.</span>
      <span v-for="plugin in plugins" :key="plugin.name" v-html="plugin.credits"><br></span>
      <span><a @click="help">Help</a></span>
    </p>
  </nav>
</template>

<script>
import {mapState} from 'vuex'
import auth from '@/utils/auth'
import buttons from '@/utils/buttons'
import api from '@/utils/api'

export default {
  name: 'sidebar',
  data: function () {
    return {
      pluginData: {
        api,
        buttons,
        'store': this.$store,
        'router': this.$router
      }
    }
  },
  computed: {
    ...mapState(['user', 'plugins']),
    active () {
      return this.$store.state.show === 'sidebar'
    }
  },
  methods: {
    help: function () {
      this.$store.commit('showHover', 'help')
    },
    logout: auth.logout
  }
}
</script>
