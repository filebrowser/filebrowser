<template>
  <header>
    <div>
      <button @click="openSidebar" :aria-label="$t('buttons.toggleSidebar')" :title="$t('buttons.toggleSidebar')" class="action">
        <i class="material-icons">menu</i>
      </button>
      <img src="../assets/logo.svg" alt="File Browser">
      <search></search>
    </div>
    <div>
      <button @click="openSearch" :aria-label="$t('buttons.search')" :title="$t('buttons.search')" class="search-button action">
        <i class="material-icons">search</i>
      </button>

      <button v-show="showSaveButton" :aria-label="$t('buttons.save')" :title="$t('buttons.save')" class="action" id="save-button">
        <i class="material-icons">save</i>
      </button>

      <template v-if="staticGen.length > 0">
        <button v-show="showPublishButton" :aria-label="$t('buttons.publish')" :title="$t('buttons.publish')" class="action" id="publish-button">
          <i class="material-icons">send</i>
        </button>
      </template>

      <button @click="openMore" id="more" :aria-label="$t('buttons.more')" :title="$t('buttons.more')" class="action">
        <i class="material-icons">more_vert</i>
      </button>

      <!-- Menu that shows on listing AND mobile when there are files selected -->
      <div id="file-selection" v-if="isMobile && req.kind === 'listing'">
        <span v-if="selectedCount > 0">{{ selectedCount }} selected</span>
        <share-button v-show="showRenameButton"></share-button>
        <rename-button v-show="showRenameButton"></rename-button>
        <copy-button v-show="showMoveButton"></copy-button>
        <move-button v-show="showMoveButton"></move-button>
        <delete-button v-show="showDeleteButton"></delete-button>
      </div>

      <!-- This buttons are shown on a dropdown on mobile phones -->
      <div id="dropdown" :class="{ active: showMore }">
        <div v-if="!isListing || !isMobile">
          <share-button v-show="showRenameButton"></share-button>
          <rename-button v-show="showRenameButton"></rename-button>
          <copy-button v-show="showMoveButton"></copy-button>
          <move-button v-show="showMoveButton"></move-button>
          <delete-button v-show="showDeleteButton"></delete-button>
        </div>

        <template v-if="staticGen.length > 0">
          <schedule-button v-show="showPublishButton"></schedule-button>
        </template>

        <switch-button v-show="showSwitchButton"></switch-button>
        <download-button v-show="showCommonButton"></download-button>
        <upload-button v-show="showUpload"></upload-button>
        <info-button v-show="showCommonButton"></info-button>

        <button v-show="showSelectButton" @click="openSelect" :aria-label="$t('buttons.selectMultiple')" :title="$t('buttons.selectMultiple')" class="action">
          <i class="material-icons">check_circle</i>
          <span>{{ $t('buttons.select') }}</span>
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
import CopyButton from './buttons/Copy'
import ScheduleButton from './buttons/Schedule'
import ShareButton from './buttons/Share'
import {mapGetters, mapState} from 'vuex'
import * as api from '@/utils/api'
import buttons from '@/utils/buttons'

export default {
  name: 'main',
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
    ScheduleButton
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
      'staticGen'
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
    showPublishButton () {
      return (this.req.kind === 'editor' && !this.loading && this.user.allowPublish)
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
