<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t('prompts.copy') }}</h2>
    </div>

    <div class="card-content">
      <p>{{ $t('prompts.copyMessage') }}</p>
      <file-list @update:selected="val => dest = val" />
    </div>

    <div class="card-action">
      <button
        class="button button--flat button--grey"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')"
        @click="$store.commit('closeHovers')"
      >{{ $t('buttons.cancel') }}</button>
      <button
        class="button button--flat"
        :disabled="$route.path === dest"
        :aria-label="$t('buttons.copy')"
        :title="$t('buttons.copy')"
        @click="copy"
      >{{ $t('buttons.copy') }}</button>
    </div>
  </div>
</template>

<script>
import { mapState } from 'vuex'
import FileList from './FileList'
import { files as api } from '@/api'
import buttons from '@/utils/buttons'

export default {
  name: 'Copy',
  components: { FileList },
  data: function() {
    return {
      current: window.location.pathname,
      dest: null
    }
  },
  computed: mapState(['req', 'selected']),
  methods: {
    copy: async function(event) {
      event.preventDefault()
      buttons.loading('copy')
      const items = []

      // Create a new promise for each file.
      for (const item of this.selected) {
        items.push({
          from: this.req.items[item].url,
          to: this.dest + encodeURIComponent(this.req.items[item].name)
        })
      }

      try {
        await api.copy(items)
        buttons.success('copy')
        this.$router.push({ path: this.dest })
      } catch (e) {
        buttons.done('copy')
        this.$showError(e)
      }
    }
  }
}
</script>
