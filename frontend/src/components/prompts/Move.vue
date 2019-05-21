<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t('prompts.move') }}</h2>
    </div>

    <div class="card-content">
      <file-list @update:selected="val => dest = val"></file-list>
    </div>

    <div class="card-action">
      <button class="button button--flat button--grey"
        @click="$store.commit('closeHovers')"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')">{{ $t('buttons.cancel') }}</button>
      <button class="button button--flat"
        @click="move"
        :disabled="$route.path === dest"
        :aria-label="$t('buttons.move')"
        :title="$t('buttons.move')">{{ $t('buttons.move') }}</button>
    </div>
  </div>
</template>

<script>
import { mapState } from 'vuex'
import FileList from './FileList'
import { files as api } from '@/api'
import buttons from '@/utils/buttons'

export default {
  name: 'move',
  components: { FileList },
  data: function () {
    return {
      current: window.location.pathname,
      dest: null
    }
  },
  computed: mapState(['req', 'selected']),
  methods: {
    move: async function (event) {
      event.preventDefault()
      buttons.loading('move')
      let items = []

      for (let item of this.selected) {
        items.push({
          from: this.req.items[item].url,
          to: this.dest + encodeURIComponent(this.req.items[item].name)
        })
      }

      try {
        api.move(items)
        buttons.success('move')
        this.$router.push({ path: this.dest })
      } catch (e) {
        buttons.done('move')
        this.$showError(e)
      }

      event.preventDefault()
    }
  }
}
</script>
