<template>
  <div class="prompt">
    <h3>{{ $t('prompts.deleteTitle') }}</h3>
    <p v-show="req.kind !== 'listing'">{{ $t('prompts.deleteMessageSingle') }}</p>
    <p v-show="req.kind === 'listing'">{{ $t('prompts.deleteMessageMultiple', { count: selectedCount}) }}</p>
    <div>
      <button @click="submit" autofocus>{{ $t('buttons.delete') }}</button>
      <button @click="closeHovers" class="cancel">{{ $t('buttons.cancel') }}</button>
    </div>
  </div>
</template>

<script>
import {mapGetters, mapMutations, mapState} from 'vuex'
import api from '@/utils/api'
import url from '@/utils/url'
import buttons from '@/utils/buttons'

export default {
  name: 'delete',
  computed: {
    ...mapGetters(['selectedCount']),
    ...mapState(['req', 'selected'])
  },
  methods: {
    ...mapMutations(['closeHovers']),
    submit: function (event) {
      this.closeHovers()
      buttons.loading('delete')

      // If we are not on a listing, delete the current
      // opened file.
      if (this.req.kind !== 'listing') {
        api.delete(this.$route.path)
          .then(() => {
            buttons.done('delete')
            this.$router.push({ path: url.removeLastDir(this.$route.path) + '/' })
          })
          .catch(error => {
            buttons.done('delete')
            this.$store.commit('showError', error)
          })

        return
      }

      if (this.selectedCount === 0) {
        // This shouldn't happen...
        return
      }

      // Create the promises array and fill it with
      // the delete request for every selected file.
      let promises = []

      for (let index of this.selected) {
        promises.push(api.delete(this.req.items[index].url))
      }

      Promise.all(promises)
        .then(() => {
          buttons.done('delete')
          this.$store.commit('setReload', true)
        })
        .catch(error => {
          buttons.done('delete')
          this.$store.commit('setReload', true)
          this.$store.commit('showError', error)
        })
    }
  }
}
</script>
