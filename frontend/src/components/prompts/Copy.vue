<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t('prompts.copy') }}</h2>
    </div>

    <div class="card-content">
      <p>{{ $t('prompts.copyMessage') }}</p>
      <file-list @update:selected="val => dest = val"></file-list>
    </div>

    <div class="card-action">
      <button class="button button--flat button--grey"
        @click="$store.commit('closeHovers')"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')">{{ $t('buttons.cancel') }}</button>
      <button class="button button--flat"
        @click="copy"
        :disabled="$route.path === dest"
        :aria-label="$t('buttons.copy')"
        :title="$t('buttons.copy')">{{ $t('buttons.copy') }}</button>
    </div>
  </div>
</template>

<script>
import { mapState } from 'vuex'
import FileList from './FileList'
import { files as api } from '@/api'
import buttons from '@/utils/buttons'
import * as upload  from '@/utils/upload'

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
    copy: async function (event) {
      event.preventDefault()
      let items = []

      // Create a new promise for each file.
      for (let item of this.selected) {
        items.push({
          from: this.req.items[item].url,
          to: this.dest + encodeURIComponent(this.req.items[item].name),
          name: this.req.items[item].name
        })
      }

      let action = async (overwrite) => {
        buttons.loading('copy')

        await api.copy(items, overwrite).then(() => {
          buttons.success('copy')
          this.$router.push({ path: this.dest })
        }).catch((e) => {
          buttons.done('copy')
          this.$showError(e)
        })
      }

      let dstItems = (await api.fetch(this.dest)).items
      let conflict = upload.checkConflict(items, dstItems)

      if (conflict) {
        this.$store.commit('showHover', {
          prompt: 'replace',
          confirm: (event) => {
            event.preventDefault()
            this.$store.commit('closeHovers')
            action(true)
          }
        })

        return
      }

      action(false)
    }
  }
}
</script>
