<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t('prompts.rename') }}</h2>
    </div>

    <div class="card-content">
      <p>{{ $t('prompts.renameMessage') }} <code>{{ oldName() }}</code>:</p>
      <input autofocus type="text" @keyup.enter="submit" v-model.trim="name">
    </div>

    <div class="card-action">
      <button class="cancel flat"
        @click="$store.commit('closeHovers')"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')">{{ $t('buttons.cancel') }}</button>
      <button @click="submit"
        class="flat"
        type="submit"
        :aria-label="$t('buttons.rename')"
        :title="$t('buttons.rename')">{{ $t('buttons.rename') }}</button>
    </div>
  </div>
</template>

<script>
import { mapState } from 'vuex'
import url from '@/utils/url'
import * as api from '@/utils/api'

export default {
  name: 'rename',
  data: function () {
    return {
      name: ''
    }
  },
  created () {
    this.name = this.oldName()
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
          this.$showError(error)
        })

      this.$store.commit('closeHovers')
    }
  }
}
</script>
