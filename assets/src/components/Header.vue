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
      <rename-button v-show="showRenameButton"></rename-button>
      <move-button v-show="showMoveButton"></move-button>
      <delete-button v-show="showDeleteButton"></delete-button>
      <switch-button v-show="showSwitchButton"></switch-button>
      <download-button v-show="showCommonButton"></download-button>
      <upload-button v-show="showUpload"></upload-button>
      <info-button v-show="showCommonButton"></info-button>

      <button v-show="showSelectButton" @click="$store.commit('multiple', true)" aria-label="Select multiple" class="action">
        <i class="material-icons">check_circle</i>
        <span>Select</span>
      </button>
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
  computed: {
    ...mapGetters([
      'selectedCount'
    ]),
    ...mapState([
      'req',
      'user',
      'loading',
      'reload',
      'multiple'
    ]),
    showSelectButton () {
      return this.req.kind === 'listing' && !this.loading && this.$route.name === 'Files'
    },
    showSaveButton () {
      return (this.req.kind === 'editor' && !this.loading) || this.$route.name === 'User'
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
    }
  },
  methods: {
    openSidebar () {
      this.$store.commit('showHover', 'sidebar')
    },
    openSearch () {
      this.$store.commit('showHover', 'search')
    }
  }
}
</script>
