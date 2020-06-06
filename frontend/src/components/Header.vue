<template>
  <header>
    <div>
      <button :aria-label="$t('buttons.toggleSidebar')" :title="$t('buttons.toggleSidebar')" class="action" @click="openSidebar">
        <i class="material-icons">menu</i>
      </button>
      <img :src="logoURL" alt="File Browser">
      <search v-if="isLogged" />
    </div>
    <div>
      <template v-if="isLogged">
        <button :aria-label="$t('buttons.search')" :title="$t('buttons.search')" class="search-button action" @click="openSearch">
          <i class="material-icons">search</i>
        </button>

        <button v-show="showSaveButton" id="save-button" :aria-label="$t('buttons.save')" :title="$t('buttons.save')" class="action">
          <i class="material-icons">save</i>
        </button>

        <button id="more" :aria-label="$t('buttons.more')" :title="$t('buttons.more')" class="action" @click="openMore">
          <i class="material-icons">more_vert</i>
        </button>

        <!-- Menu that shows on listing AND mobile when there are files selected -->
        <div v-if="isMobile && isListing" id="file-selection">
          <span v-if="selectedCount > 0">{{ selectedCount }} selected</span>
          <share-button v-show="showShareButton" />
          <rename-button v-show="showRenameButton" />
          <copy-button v-show="showCopyButton" />
          <move-button v-show="showMoveButton" />
          <delete-button v-show="showDeleteButton" />
        </div>

        <!-- This buttons are shown on a dropdown on mobile phones -->
        <div id="dropdown" :class="{ active: showMore }">
          <div v-if="!isListing || !isMobile">
            <share-button v-show="showShareButton" />
            <rename-button v-show="showRenameButton" />
            <copy-button v-show="showCopyButton" />
            <move-button v-show="showMoveButton" />
            <delete-button v-show="showDeleteButton" />
          </div>

          <shell-button v-show="user.perm.execute" />
          <switch-button v-show="isListing" />
          <download-button v-show="showDownloadButton" />
          <upload-button v-show="showUpload" />
          <info-button v-show="isFiles" />

          <button v-show="isListing" :aria-label="$t('buttons.selectMultiple')" :title="$t('buttons.selectMultiple')" class="action" @click="toggleMultipleSelection">
            <i class="material-icons">check_circle</i>
            <span>{{ $t('buttons.select') }}</span>
          </button>
        </div>
      </template>

      <div v-show="showOverlay" class="overlay" @click="resetPrompts" />
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
import CopyButton from './buttons/Copy'
import ShareButton from './buttons/Share'
import ShellButton from './buttons/Shell'
import { mapGetters, mapState } from 'vuex'
import { logoURL } from '@/utils/constants'
import * as api from '@/api'
import buttons from '@/utils/buttons'

export default {
  name: 'HeaderLayout',
  components: {
    Search,
    InfoButton,
    DeleteButton,
    ShareButton,
    RenameButton,
    DownloadButton,
    CopyButton,
    UploadButton,
    SwitchButton,
    MoveButton,
    ShellButton
  },
  data: function() {
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
  created() {
    window.addEventListener('resize', () => {
      this.width = window.innerWidth
    })
  },
  computed: {
    ...mapGetters([
      'selectedCount',
      'isFiles',
      'isEditor',
      'isListing',
      'isLogged'
    ]),
    ...mapState([
      'req',
      'user',
      'loading',
      'reload',
      'multiple'
    ]),
    logoURL: () => logoURL,
    isMobile() {
      return this.width <= 736
    },
    showUpload() {
      return this.isListing && this.user.perm.create
    },
    showSaveButton() {
      return this.isEditor && this.user.perm.modify
    },
    showDownloadButton() {
      return this.isFiles && this.user.perm.download
    },
    showDeleteButton() {
      return this.isFiles && (this.isListing
        ? (this.selectedCount !== 0 && this.user.perm.delete)
        : this.user.perm.delete)
    },
    showRenameButton() {
      return this.isFiles && (this.isListing
        ? (this.selectedCount === 1 && this.user.perm.rename)
        : this.user.perm.rename)
    },
    showShareButton() {
      return this.isFiles && (this.isListing
        ? (this.selectedCount === 1 && this.user.perm.share)
        : this.user.perm.share)
    },
    showMoveButton() {
      return this.isFiles && (this.isListing
        ? (this.selectedCount > 0 && this.user.perm.rename)
        : this.user.perm.rename)
    },
    showCopyButton() {
      return this.isFiles && (this.isListing
        ? (this.selectedCount > 0 && this.user.perm.create)
        : this.user.perm.create)
    },
    showMore() {
      return this.isFiles && this.$store.state.show === 'more'
    },
    showOverlay() {
      return this.showMore
    }
  },
  methods: {
    openSidebar() {
      this.$store.commit('showHover', 'sidebar')
    },
    openMore() {
      this.$store.commit('showHover', 'more')
    },
    openSearch() {
      this.$store.commit('showHover', 'search')
    },
    toggleMultipleSelection() {
      this.$store.commit('multiple', !this.multiple)
      this.resetPrompts()
    },
    resetPrompts() {
      this.$store.commit('closeHovers')
    }
  }
}
</script>
