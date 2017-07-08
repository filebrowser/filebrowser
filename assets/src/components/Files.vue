<template>
  <div v-if="error">
    <not-found v-if="error === 404"></not-found>
    <forbidden v-else-if="error === 403"></forbidden>
    <internal-error v-else></internal-error>
  </div>
  <editor v-else-if="isEditor"></editor>
  <listing :class="{ multiple }" v-else-if="isListing"></listing>
  <preview v-else-if="isPreview"></preview>
</template>

<script>
import Forbidden from './errors/403'
import NotFound from './errors/404'
import InternalError from './errors/500'
import Preview from './Preview'
import Listing from './Listing'
import Editor from './Editor'
import css from '@/utils/css'
import api from '@/utils/api'
import { mapGetters, mapState, mapMutations } from 'vuex'

function updateColumnSizes () {
  let columns = Math.floor(document.querySelector('main').offsetWidth / 300)
  let items = css(['#listing.mosaic .item', '.mosaic#listing .item'])

  if (columns === 0) columns = 1

  items.style.width = `calc(${100 / columns}% - 1em)`
}

export default {
  name: 'files',
  components: {
    Forbidden,
    NotFound,
    InternalError,
    Preview,
    Listing,
    Editor
  },
  computed: {
    ...mapGetters([
      'selectedCount'
    ]),
    ...mapState([
      'req',
      'user',
      'reload',
      'multiple',
      'loading'
    ]),
    isListing () {
      return this.req.kind === 'listing' && !this.loading
    },
    isPreview () {
      return this.req.kind === 'preview' && !this.loading
    },
    isEditor () {
      return this.req.kind === 'editor' && !this.loading
    }
  },
  data: function () {
    return {
      error: null
    }
  },
  created () {
    this.fetchData()
  },
  watch: {
    '$route': 'fetchData',
    'reload': function () {
      this.$store.commit('setReload', false)
      this.fetchData()
    }
  },
  mounted () {
    updateColumnSizes()
    window.addEventListener('resize', updateColumnSizes)
    window.addEventListener('keydown', this.keyEvent)
  },
  methods: {
    ...mapMutations([ 'setLoading' ]),
    fetchData () {
      // Set loading to true and reset the error.
      this.setLoading(true)
      this.error = null

      let url = this.$route.path
      if (url === '') url = '/'
      if (url[0] !== '/') url = '/' + url

      api.fetch(url)
      .then((req) => {
        if (!url.endsWith('/') && req.url.endsWith('/')) {
          window.history.replaceState(window.history.state, document.title, window.location.pathname + '/')
        }

        this.$store.commit('updateRequest', req)
        document.title = req.name
        this.setLoading(false)
      })
      .catch(error => {
        this.setLoading(false)

        if (typeof error === 'object') {
          this.error = error.status
          return
        }

        this.error = error
      })
    },
    keyEvent (event) {
      // Esc!
      if (event.keyCode === 27) {
        this.$store.commit('closeHovers')

        if (this.req.kind !== 'listing') {
          return
        }

        // If we're on a listing, unselect all files and folders.
        let items = document.getElementsByClassName('item')
        Array.from(items).forEach(link => {
          link.setAttribute('aria-selected', false)
        })

        this.$store.commit('resetSelected')
      }

      // Del!
      if (event.keyCode === 46) {
        if (this.req.kind === 'editor' ||
          this.$route.name !== 'Files' ||
          this.loading ||
          !this.user.allowEdit ||
          (this.req.kind === 'listing' && this.selectedCount === 0)) return

        this.$store.commit('showHover', 'delete')
      }

      // F1!
      if (event.keyCode === 112) {
        event.preventDefault()
        this.$store.commit('showHover', 'help')
      }

      // F2!
      if (event.keyCode === 113) {
        if (this.req.kind === 'editor' ||
          this.$route.name !== 'Files' ||
          this.loading ||
          !this.user.allowEdit ||
          (this.req.kind === 'listing' && this.selectedCount === 0) ||
          (this.req.kind === 'listing' && this.selectedCount > 1)) return

        this.$store.commit('showHover', 'rename')
      }

      // CTRL + S
      if (event.ctrlKey || event.metaKey) {
        if (String.fromCharCode(event.which).toLowerCase() === 's') {
          event.preventDefault()

          if (this.req.kind !== 'editor') {
            document.getElementById('download-button').click()
            return
          }
        }
      }
    },
    openSidebar () {
      this.$store.commit('showHover', 'sidebar')
    },
    openSearch () {
      this.$store.commit('showHover', 'search')
    }
  }
}
</script>
