<template>
  <nav>
    <router-link class="action" to="/files/" aria-label="My Files" title="My Files">
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

    <div>
      <router-link class="action" to="/dashboard" aria-label="Settings" title="Settings">
        <i class="material-icons">settings_applications</i>
        <span>Settings</span>
      </router-link>

      <button @click="logout" class="action" id="logout" aria-label="Log out" title="Logout">
        <i class="material-icons">exit_to_app</i>
        <span>Logout</span>
      </button>
    </div>

    <p class="credits">Served with <a rel="noopener noreferrer" href="https://github.com/hacdias/caddy-filemanager">File Manager</a>.</p>
  </nav>
</template>

<script>
import {mapState} from 'vuex'
import auth from '@/utils/auth'

export default {
  name: 'sidebar',
  data: () => {
    return {
      plugins: []
    }
  },
  computed: mapState(['user']),
  mounted () {
    if (window.plugins !== undefined || window.plugins !== null) {
      this.plugins = window.plugins
    }
  },
  methods: {
    logout: auth.logout
  }
}
</script>
