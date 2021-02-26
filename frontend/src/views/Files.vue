<template>
  <div>
    <header-bar showMenu showLogo />

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
    <listing v-else-if="isListing"></listing>
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

import HeaderBar from '@/components/header/HeaderBar'
import Errors from '@/views/Errors'
import Preview from '@/views/files/Preview'
import Listing from '@/views/files/Listing'

function clean (path) {
  return path.endsWith('/') ? path.slice(0, -1) : path
}

export default {
  name: 'files',
  components: {
    HeaderBar,
    Errors,
    Preview,
    Listing,
    Editor: () => import('@/views/files/Editor'),
  },
  data: function () {
    return {
      error: null,
      width: window.innerWidth
    }
  },
  computed: {
    ...mapGetters([
      'isListing',
      'isEditor',
      'isFiles'
    ]),
    ...mapState([
      'req',
      'reload',
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
    errorCode() {
      return (this.error.message === '404' || this.error.message === '403') ? parseInt(this.error.message) : 500
    }
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
  },
  beforeDestroy () {
    window.removeEventListener('keydown', this.keyEvent)
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

      // F1!
      if (event.keyCode === 112) {
        event.preventDefault()
        this.$store.commit('showHover', 'help')
      }
    }
  }
}
</script>
