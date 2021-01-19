<template>
  <div class="card floating">
    <div class="card-content">
      <p>{{ $t('prompts.deleteMessageShare', {path: hash.path}) }}</p>
    </div>
    <div class="card-action">
      <button @click="$store.commit('closeHovers')"
        class="button button--flat button--grey"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')">{{ $t('buttons.cancel') }}</button>
      <button @click="submit"
        class="button button--flat button--red"
        :aria-label="$t('buttons.delete')"
        :title="$t('buttons.delete')">{{ $t('buttons.delete') }}</button>
    </div>
  </div>
</template>

<script>
import {mapMutations, mapState} from 'vuex'
import { share as api } from '@/api'
import buttons from '@/utils/buttons'

export default {
  name: 'share-delete',
  computed: {
    ...mapState(['hash'])
  },
  methods: {
    ...mapMutations(['closeHovers']),
    submit: async function () {
      buttons.loading('delete')

      try {
        await api.remove(this.hash.hash)
        buttons.success('delete')

        this.$root.$emit('share-deleted', this.hash.hash)
        this.closeHovers()
      } catch (e) {
        buttons.done('delete')
        this.$showError(e)
      }
    }
  }
}
</script>
