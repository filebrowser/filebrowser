<template>
  <nav :class="{active}">
    <router-link class="action" to="/files/" :aria-label="$t('sidebar.myFiles')" :title="$t('sidebar.myFiles')">
      <i class="material-icons">folder</i>
      <span>{{ $t('sidebar.myFiles') }}</span>
    </router-link>

    <div v-if="user.allowNew">
      <button @click="$store.commit('showHover', 'newDir')" class="action" :aria-label="$t('sidebar.newFolder')" :title="$t('sidebar.newFolder')">
        <i class="material-icons">create_new_folder</i>
        <span>{{ $t('sidebar.newFolder') }}</span>
      </button>

      <button @click="$store.commit('showHover', 'newFile')" class="action" :aria-label="$t('sidebar.newFile')" :title="$t('sidebar.newFile')">
        <i class="material-icons">note_add</i>
        <span>{{ $t('sidebar.newFile') }}</span>
      </button>
    </div>

    <div v-for="plugin in plugins" :key="plugin.name">
      <button v-for="action in plugin.sidebar" @click="action.click($event, pluginData, $route)" :aria-label="action.name" :title="action.name" :key="action.name" class="action">
        <i class="material-icons">{{ action.icon }}</i>
        <span>{{ action.name }}</span>
      </button>
    </div>

    <div>
      <router-link class="action" to="/settings" :aria-label="$t('sidebar.settings')" :title="$t('sidebar.settings')">
        <i class="material-icons">settings_applications</i>
        <span>{{ $t('sidebar.settings') }}</span>
      </router-link>

      <button @click="logout" class="action" id="logout" :aria-label="$t('sidebar.logout')" :title="$t('sidebar.logout')">
        <i class="material-icons">exit_to_app</i>
        <span>{{ $t('sidebar.logout') }}</span>
      </button>
    </div>

    <p class="credits">
      <span>{{ $t('sidebar.servedWith') }} <a rel="noopener noreferrer" href="https://github.com/hacdias/filemanager">File Manager</a>.</span>
      <span v-for="plugin in plugins" :key="plugin.name" v-html="plugin.credits"><br></span>
      <span><a @click="help">{{ $t('sidebar.help') }}</a></span>
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
