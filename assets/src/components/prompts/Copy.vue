<template>
  <div class="prompt">
    <h3>{{ $t('prompts.copy') }}</h3>
    <p>{{ $t('prompts.copyMessage') }}</p>

    <file-list @update:selected="val => dest = val"></file-list>

    <div>
      <button class="ok"
        @click="copy"
        :aria-label="$t('buttons.copy')"
        :title="$t('buttons.copy')">{{ $t('buttons.copy') }}</button>
      <button class="cancel"
        @click="$store.commit('closeHovers')"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')">{{ $t('buttons.cancel') }}</button>
    </div>
  </div>
</template>

<script>
import { mapState } from 'vuex'
import FileList from './FileList'
import * as api from '@/utils/api'
import buttons from '@/utils/buttons'

export default {
  name: 'copy',
  components: { FileList },
  data: function () {
    return {
      current: window.location.pathname,
      dest: null
    }
  },
  computed: mapState(['req', 'selected']),
  methods: {
    copy: function (event) {
      event.preventDefault()
      buttons.loading('copy')
      let items = []

      // Create a new promise for each file.
      for (let item of this.selected) {
        items.push({
          from: this.req.items[item].url,
          to: this.dest + encodeURIComponent(this.req.items[item].name)
        })
      }

      // Execute the promises.
      api.copy(items)
        .then(() => {
          buttons.success('copy')
          this.$router.push({ path: this.dest })
        })
        .catch(error => {
          buttons.done('copy')
          this.$store.commit('showError', error)
        })
    }
  }
}
</script>
