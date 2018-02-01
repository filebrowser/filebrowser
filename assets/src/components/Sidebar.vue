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

    <div v-if="staticGen.length > 0">
      <router-link to="/files/settings"
        :aria-label="$t('sidebar.siteSettings')"
        :title="$t('sidebar.siteSettings')"
        class="action">
        <i class="material-icons">settings</i>
        <span>{{ $t('sidebar.siteSettings') }}</span>
      </router-link>

      <template v-if="staticGen === 'hugo'">
        <button class="action"
          :aria-label="$t('sidebar.hugoNew')"
          :title="$t('sidebar.hugoNew')"
          v-if="user.allowNew"
          @click="$store.commit('showHover', 'new-archetype')">
          <i class="material-icons">merge_type</i>
          <span>{{ $t('sidebar.hugoNew') }}</span>
        </button>
      </template>

      <button class="action"
        :aria-label="$t('sidebar.preview')"
        :title="$t('sidebar.preview')"
        @click="preview">
        <i class="material-icons">remove_red_eye</i>
        <span>{{ $t('sidebar.preview') }}</span>
      </button>
    </div>

    <div v-if="!$store.state.noAuth">
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
      <span><a rel="noopener noreferrer" href="https://github.com/filebrowser/filebrowser">File Browser</a> v{{ version }}</span>
      <span><a @click="help">{{ $t('sidebar.help') }}</a></span>
    </p>
  </nav>
</template>

<script>
import {mapState} from 'vuex'
import auth from '@/utils/auth'

export default {
  name: 'sidebar',
  computed: {
    ...mapState(['user', 'staticGen', 'version']),
    active () {
      return this.$store.state.show === 'sidebar'
    }
  },
  methods: {
    help () {
      this.$store.commit('showHover', 'help')
    },
    preview () {
      window.open(this.$store.state.baseURL + '/preview/')
    },
    logout: auth.logout
  }
}
</script>
