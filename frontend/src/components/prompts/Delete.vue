<template>
  <div class="card floating">
    <div class="card-content">
      <p v-if="req.kind !== 'listing'">{{ $t('prompts.deleteMessageSingle') }}</p>
      <p v-else>{{ $t('prompts.deleteMessageMultiple', { count: selectedCount}) }}</p>
    </div>
    <div class="card-action">
      <button
        class="button button--flat button--grey"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')"
        @click="$store.commit('closeHovers')"
      >{{ $t('buttons.cancel') }}</button>
      <button
        class="button button--flat button--red"
        :aria-label="$t('buttons.delete')"
        :title="$t('buttons.delete')"
        @click="submit"
      >{{ $t('buttons.delete') }}</button>
    </div>
  </div>
</template>

<script>
import { mapGetters, mapMutations, mapState } from 'vuex'
import { files as api } from '@/api'
import url from '@/utils/url'
import buttons from '@/utils/buttons'

export default {
  name: 'Delete',
  computed: {
    ...mapGetters(['isListing', 'selectedCount']),
    ...mapState(['req', 'selected'])
  },
  methods: {
    ...mapMutations(['closeHovers']),
    submit: async function() {
      this.closeHovers()
      buttons.loading('delete')

      try {
        if (!this.isListing) {
          await api.remove(this.$route.path)
          buttons.success('delete')
          this.$router.push({ path: url.removeLastDir(this.$route.path) + '/' })
          return
        }

        if (this.selectedCount === 0) {
          return
        }

        const promises = []
        for (const index of this.selected) {
          promises.push(api.remove(this.req.items[index].url))
        }

        await Promise.all(promises)
        buttons.success('delete')
        this.$store.commit('setReload', true)
      } catch (e) {
        buttons.done('delete')
        this.$showError(e)
        if (this.isListing) this.$store.commit('setReload', true)
      }
    }
  }
}
</script>
