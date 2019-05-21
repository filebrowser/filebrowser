<template>
  <button @click="toggle" :aria-label="$t('buttons.bookmark')" :title="$t('buttons.bookmark')" class="action" id="bookmark-button">
    <i class="material-icons">{{ icon }}</i>
    <span>{{ $t('buttons.bookmark') }}</span>
  </button>
</template>

<script>
import { mapState, mapMutations } from 'vuex'
import { files as filesApi, context as contextApi } from '@/api'

export default {
  name: 'bookmark-button',
  computed: {
    ...mapState(['req']),
    icon: function () {
      if (this.req.bookmarked) return 'bookmark'
      return 'bookmark_border'
    }
  },
  methods: {
    ...mapMutations([ 'updateBookmark', 'closeHovers' ]),
    toggle: async function () {
      this.closeHovers()

      const data = {
        path:       this.req.path,
        bookmarked: (this.icon === 'bookmark_border')
      }

      try {
        await filesApi.bookmark([data])
        this.updateBookmark(data)

        const ctx = await contextApi.get()
        this.$store.commit('updateContext', ctx)
      } catch (e) {
        this.$showError(e)
      }
    }
  }
}
</script>
