<template>
  <div class="prompt">
    <h3>{{ $t('prompts.rename') }}</h3>
    <p>{{ $t('prompts.renameMessage') }} <code>{{ oldName() }}</code>:</p>

    <input autofocus type="text" @keyup.enter="submit" v-model.trim="name">
    <div>
      <button @click="submit" type="submit">{{ $t('buttons.rename') }}</button>
      <button @click="cancel" class="cancel">{{ $t('buttons.cancel') }}</button>
    </div>
  </div>
</template>

<script>
import { mapState } from 'vuex'
import url from '@/utils/url'
import api from '@/utils/api'

export default {
  name: 'rename',
  data: function () {
    return {
      name: ''
    }
  },
  computed: mapState(['req', 'selected', 'selectedCount']),
  methods: {
    cancel: function (event) {
      this.$store.commit('closeHovers')
    },
    oldName: function () {
      // Get the current name of the file we are editing.
      if (this.req.kind !== 'listing') {
        return this.req.name
      }

      if (this.selectedCount === 0 || this.selectedCount > 1) {
        // This shouldn't happen.
        return
      }

      return this.req.items[this.selected[0]].name
    },
    submit: function (event) {
      let oldLink = ''
      let newLink = ''

      if (this.req.kind !== 'listing') {
        oldLink = this.req.url
      } else {
        oldLink = this.req.items[this.selected[0]].url
      }

      this.name = encodeURIComponent(this.name)
      newLink = url.removeLastDir(oldLink) + '/' + this.name

      api.move([{ from: oldLink, to: newLink }])
        .then(() => {
          if (this.req.kind !== 'listing') {
            this.$router.push({ path: newLink })
            return
          }
          this.$store.commit('setReload', true)
        }).catch(error => {
          this.$store.commit('showError', error)
        })

      this.$store.commit('closeHovers')
    }
  }
}
</script>
