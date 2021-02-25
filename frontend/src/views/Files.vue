<template>
  <div>
    <header-bar showMenu showLogo>
      <search /> <title />
      <action class="search-button" icon="search" :label="$t('buttons.search')" @action="openSearch()" />

      <template #actions v-if="!error">
        <template v-if="!isMobile">
          <share-button v-if="headerButtons.share" />
          <rename-button v-if="headerButtons.rename" />
          <copy-button v-if="headerButtons.copy" />
          <move-button v-if="headerButtons.move" />
          <delete-button v-if="headerButtons.delete" />
        </template>

        <shell-button v-if="headerButtons.shell" />
        <switch-button />
        <download-button v-if="headerButtons.download" />
        <upload-button v-if="headerButtons.upload" />
        <info-button />
        <action icon="check_circle" :label="$t('buttons.selectMultiple')" @action="toggleMultipleSelection" />
      </template>
    </header-bar>

    <div id="file-selection" v-if="isMobile">
      <span v-if="selectedCount > 0">{{ selectedCount }} selected</span>
      <share-button v-if="headerButtons.share" />
      <rename-button v-if="headerButtons.rename" />
      <copy-button v-if="headerButtons.copy" />
      <move-button v-if="headerButtons.move" />
      <delete-button v-if="headerButtons.delete" />
    </div>

    <div id="breadcrumbs" v-if="isListing || error">
      <router-link to="/files/" :aria-label="$t('files.home')" :title="$t('files.home')">
        <i class="material-icons">home</i>
      </router-link>

      <span v-for="(link, index) in breadcrumbs" :key="index">
        <span class="chevron"><i class="material-icons">keyboard_arrow_right</i></span>
        <router-link :to="link.url">{{ link.name }}</router-link>
      </span>
    </div>

    <errors v-if="error" :errorCode="errorCode" />
    <preview v-else-if="isPreview"></preview>
    <editor v-else-if="isEditor"></editor>
    <listing :class="{ multiple }" v-else-if="isListing"></listing>
    <div v-else>
      <h2 class="message">
        <span>{{ $t('files.loading') }}</span>
      </h2>
    </div>
  </div>
</template>

<script>
import { files as api } from '@/api'
import { mapGetters, mapState, mapMutations } from 'vuex'
import { enableExec } from '@/utils/constants'

import HeaderBar from '@/components/header/HeaderBar'
import Action from '@/components/header/Action'
import Search from '@/components/Search'
import InfoButton from '@/components/buttons/Info'
import DeleteButton from '@/components/buttons/Delete'
import RenameButton from '@/components/buttons/Rename'
import UploadButton from '@/components/buttons/Upload'
import DownloadButton from '@/components/buttons/Download'
import SwitchButton from '@/components/buttons/SwitchView'
import MoveButton from '@/components/buttons/Move'
import CopyButton from '@/components/buttons/Copy'
import ShareButton from '@/components/buttons/Share'
import ShellButton from '@/components/buttons/Shell'

import Errors from '@/views/Errors'
import Preview from '@/components/files/Preview'
import Listing from '@/components/files/Listing'

function clean (path) {
  return path.endsWith('/') ? path.slice(0, -1) : path
}

export default {
  name: 'files',
  components: {
    HeaderBar,
    Action,
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
    ShellButton,
    Errors,
    Preview,
    Listing,
    Editor: () => import('@/components/files/Editor'),
  },
  data: function () {
    return {
      error: null,
      width: window.innerWidth
    }
  },
  computed: {
    ...mapGetters([
      'selectedCount',
      'isListing',
      'isEditor',
      'isFiles'
    ]),
    ...mapState([
      'req',
      'user',
      'reload',
      'multiple',
      'loading',
      'show'
    ]),
    isPreview () {
      return !this.loading && !this.isListing && !this.isEditor || this.loading && this.$store.state.previewMode
    },
    breadcrumbs () {
      let parts = this.$route.path.split('/')

      if (parts[0] === '') {
        parts.shift()
      }

      if (parts[parts.length - 1] === '') {
        parts.pop()
      }

      let breadcrumbs = []

      for (let i = 0; i < parts.length; i++) {
        if (i === 0) {
          breadcrumbs.push({ name: decodeURIComponent(parts[i]), url: '/' + parts[i] + '/' })
        } else {
          breadcrumbs.push({ name: decodeURIComponent(parts[i]), url: breadcrumbs[i - 1].url + parts[i] + '/' })
        }
      }

      breadcrumbs.shift()

      if (breadcrumbs.length > 3) {
        while (breadcrumbs.length !== 4) {
          breadcrumbs.shift()
        }

        breadcrumbs[0].name = '...'
      }

      return breadcrumbs
    },
    headerButtons() {
      return {
        upload: this.user.perm.create,
        download: this.user.perm.download,
        shell: this.user.perm.execute && enableExec,
        delete: this.selectedCount > 0 && this.user.perm.delete,
        rename: this.selectedCount === 1 && this.user.perm.rename,
        share: this.selectedCount === 1 && this.user.perm.share,
        move: this.selectedCount === 1 && this.user.perm.rename,
        copy: this.selectedCount === 1 && this.user.perm.create,
      }
    },
    errorCode() {
      return (this.error.message === '404' || this.error.message === '403') ? parseInt(this.error.message) : 500
    },
    isMobile () {
      return this.width <= 736
    },
  },
  created () {
    this.fetchData()
  },
  watch: {
    '$route': 'fetchData',
    'reload': function (value) {
       if (value === true) {
        this.fetchData()
      }
    }
  },
  mounted () {
    window.addEventListener('keydown', this.keyEvent)
    window.addEventListener('scroll', this.scroll)
    window.addEventListener('resize', this.windowsResize)
  },
  beforeDestroy () {
    window.removeEventListener('keydown', this.keyEvent)
    window.removeEventListener('scroll', this.scroll)
    window.removeEventListener('resize', this.windowsResize)
  },
  destroyed () {
    if (this.$store.state.showShell) {
      this.$store.commit('toggleShell')
    }
    this.$store.commit('updateRequest', {})
  },
  methods: {
    ...mapMutations([ 'setLoading' ]),
    async fetchData () {
      // Reset view information.
      this.$store.commit('setReload', false)
      this.$store.commit('resetSelected')
      this.$store.commit('multiple', false)
      this.$store.commit('closeHovers')

      // Set loading to true and reset the error.
      this.setLoading(true)
      this.error = null

      let url = this.$route.path
      if (url === '') url = '/'
      if (url[0] !== '/') url = '/' + url

      try {
        const res = await api.fetch(url)

        if (clean(res.path) !== clean(`/${this.$route.params.pathMatch}`)) {
          return
        }

        this.$store.commit('updateRequest', res)
        document.title = res.name
      } catch (e) {
        this.error = e
      } finally {
        this.setLoading(false)
      }
    },
    keyEvent (event) {
      if (this.show !== null) {
        // Esc!
        if (event.keyCode === 27) {
          this.$store.commit('closeHovers')
        }

        return
      }

      // Esc!
      if (event.keyCode === 27) {
        // If we're on a listing, unselect all
        // files and folders.
        if (this.isListing) {
          this.$store.commit('resetSelected')
        }
      }

      // Del!
      if (event.keyCode === 46) {
        if (this.isEditor ||
          !this.isFiles ||
          this.loading ||
          !this.user.perm.delete ||
          (this.isListing && this.selectedCount === 0) ||
          this.$store.state.show != null) return

        this.$store.commit('showHover', 'delete')
      }

      // F1!
      if (event.keyCode === 112) {
        event.preventDefault()
        this.$store.commit('showHover', 'help')
      }

      // F2!
      if (event.keyCode === 113) {
        if (this.isEditor ||
          !this.isFiles ||
          this.loading ||
          !this.user.perm.rename ||
          (this.isListing && this.selectedCount === 0) ||
          (this.isListing && this.selectedCount > 1)) return

        this.$store.commit('showHover', 'rename')
      }

      // CTRL + S
      if (event.ctrlKey || event.metaKey) {
        if (this.isEditor) return

        if (String.fromCharCode(event.which).toLowerCase() === 's') {
          event.preventDefault()

          if (this.req.kind !== 'editor') {
            document.getElementById('download-button').click()
          }
        }
      }
    },
    scroll () {
      if (this.req.kind !== 'listing' || this.$store.state.user.viewMode === 'mosaic') return

      let top = 112 - window.scrollY

      if (top < 64) {
        top = 64
      }

      document.querySelector('#listing.list .item.header').style.top = top + 'px'
    },
    openSearch () {
      this.$store.commit('showHover', 'search')
    },
    toggleMultipleSelection () {
      this.$store.commit('multiple', !this.multiple)
      this.$store.commit('closeHovers')
    },
    windowsResize () {
      this.width = window.innerWidth
    }
  }
}
</script>
