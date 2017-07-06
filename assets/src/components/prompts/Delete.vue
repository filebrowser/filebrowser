<template>
  <div class="prompt">
    <h3>Delete files</h3>
    <p v-show="req.kind !== 'listing'">Are you sure you want to delete this file/folder?</p>
    <p v-show="req.kind === 'listing'">Are you sure you want to delete {{ selectedCount }} file(s)?</p>
    <div>
      <button @click="submit" autofocus>Delete</button>
      <button @click="closeHovers" class="cancel">Cancel</button>
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

      if (this.req.kind !== 'listing') {
        api.delete(this.$route.path)
          .then(() => {
            buttons.done('delete')
            this.$router.push({path: url.removeLastDir(this.$route.path) + '/'})
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

      let promises = []

      for (let index of this.selected) {
        promises.push(api.delete(this.req.items[index].url))
      }

      Promise.all(promises)
        .then(() => {
          this.$store.commit('setReload', true)
          buttons.done('delete')
        })
        .catch(error => {
          console.log(error)
          this.$store.commit('setReload', true)
          buttons.done('delete')
          this.$store.commit('showError', error)
        })
    }
  }
}
</script>
