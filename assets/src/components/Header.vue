<template>
  <header>
    <div>
      <button @click="openSidebar" aria-label="Toggle sidebar" title="Toggle sidebar" class="action">
        <i class="material-icons">menu</i>
      </button>
      <img src="../assets/logo.svg" alt="File Manager">
      <search></search>
    </div>
    <div>
      <button @click="openSearch" aria-label="Search" title="Search" class="search-button action">
        <i class="material-icons">search</i>
      </button>

      <button v-show="showSaveButton" aria-label="Save" class="action" id="save-button">
        <i class="material-icons" title="Save">save</i>
      </button>

      <div v-for="plugin in plugins" :key="plugin.name">
        <button class="action"
          v-for="action in plugin.header.visible"
          v-if="action.if(pluginData, $route)"
          @click="action.click($event, pluginData, $route)"
          :aria-label="action.name"
          :id="action.id"
          :title="action.name"
          :key="action.name">
          <i class="material-icons">{{ action.icon }}</i>
          <span>{{ action.name }}</span>
        </button>
      </div>

      <button @click="openMore" id="more" aria-label="More" title="More" class="action">
        <i class="material-icons">more_vert</i>
      </button>

      <!-- Menu that shows on listing AND mobile when there are files selected -->
      <div id="file-selection" v-if="isMobile && req.kind === 'listing'">
        <span v-if="selectedCount > 0">{{ selectedCount }} selected</span>
        <rename-button v-show="showRenameButton"></rename-button>
        <move-button v-show="showMoveButton"></move-button>
        <delete-button v-show="showDeleteButton"></delete-button>
      </div>

      <!-- This buttons are shown on a dropdown on mobile phones -->
      <div id="dropdown" :class="{ active: showMore }">
        <div v-if="!isListing || !isMobile">
          <rename-button v-show="showRenameButton"></rename-button>
          <move-button v-show="showMoveButton"></move-button>
          <delete-button v-show="showDeleteButton"></delete-button>
        </div>

        <div v-for="plugin in plugins" :key="plugin.name">
          <button class="action"
            v-for="action in plugin.header.hidden"
            v-if="action.if(pluginData, $route)"
            @click="action.click($event, pluginData, $route)"
            :id="action.id"
            :aria-label="action.name"
            :title="action.name"
            :key="action.name">
            <i class="material-icons">{{ action.icon }}</i>
            <span>{{ action.name }}</span>
          </button>
        </div>

        <switch-button v-show="showSwitchButton"></switch-button>
        <download-button v-show="showCommonButton"></download-button>
        <upload-button v-show="showUpload"></upload-button>
        <info-button v-show="showCommonButton"></info-button>

        <button v-show="showSelectButton" @click="openSelect" aria-label="Select multiple" class="action">
          <i class="material-icons">check_circle</i>
          <span>Select</span>
        </button>
      </div>
      <div v-show="showOverlay" @click="resetPrompts" class="overlay"></div>
    </div>
  </header>
</template>

<script>
import Search from './Search'
import InfoButton from './buttons/Info'
import DeleteButton from './buttons/Delete'
import RenameButton from './buttons/Rename'
import UploadButton from './buttons/Upload'
import DownloadButton from './buttons/Download'
import SwitchButton from './buttons/SwitchView'
import MoveButton from './buttons/Move'
import {mapGetters, mapState} from 'vuex'
import api from '@/utils/api'
import buttons from '@/utils/buttons'

export default {
  name: 'main',
  components: {
    Search,
    InfoButton,
    DeleteButton,
    RenameButton,
    DownloadButton,
    UploadButton,
    SwitchButton,
    MoveButton
  },
  data: function () {
    return {
      width: window.innerWidth,
      pluginData: {
        api,
        buttons,
        'store': this.$store,
        'router': this.$router
      }
    }
  },
  created () {
    window.addEventListener('resize', () => {
      this.width = window.innerWidth
    })
  },
  computed: {
    ...mapGetters([
      'selectedCount'
    ]),
    ...mapState([
      'req',
      'user',
      'loading',
      'reload',
      'multiple',
      'plugins'
    ]),
    isMobile () {
      return this.width <= 736
    },
    isListing () {
      return this.req.kind === 'listing'
    },
    showSelectButton () {
      return this.req.kind === 'listing' && !this.loading && this.$route.name === 'Files'
    },
    showSaveButton () {
      return (this.req.kind === 'editor' && !this.loading)
    },
    showSwitchButton () {
      return this.req.kind === 'listing' && this.$route.name === 'Files' && !this.loading
    },
    showCommonButton () {
      return !(this.$route.name !== 'Files' || this.loading)
    },
    showUpload () {
      if (this.$route.name !== 'Files' || this.loading) return false

      if (this.req.kind === 'editor') return false
      return this.user.allowNew
    },
    showDeleteButton () {
      if (this.$route.name !== 'Files' || this.loading) return false

      if (this.req.kind === 'listing') {
        if (this.selectedCount === 0) {
          return false
        }

        return this.user.allowEdit
      }

      return this.user.allowEdit
    },
    showRenameButton () {
      if (this.$route.name !== 'Files' || this.loading) return false

      if (this.req.kind === 'listing') {
        if (this.selectedCount === 1) {
          return this.user.allowEdit
        }

        return false
      }

      return this.user.allowEdit
    },
    showMoveButton () {
      if (this.$route.name !== 'Files' || this.loading) return false

      if (this.req.kind !== 'listing') {
        return false
      }

      if (this.selectedCount > 0) {
        return this.user.allowEdit
      }

      return false
    },
    showMore () {
      if (this.$route.name !== 'Files' || this.loading) return false
      return (this.$store.state.show === 'more')
    },
    showOverlay () {
      return (this.$store.state.show === 'more')
    }
  },
  methods: {
    openSidebar () {
      this.$store.commit('showHover', 'sidebar')
    },
    openMore () {
      this.$store.commit('showHover', 'more')
    },
    openSearch () {
      this.$store.commit('showHover', 'search')
    },
    openSelect () {
      this.$store.commit('multiple', true)
      this.resetPrompts()
    },
    resetPrompts () {
      this.$store.commit('closeHovers')
    }
  }
}
</script>
