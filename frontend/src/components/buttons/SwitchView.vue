<template>
  <button id="switch-view-button" :aria-label="$t('buttons.switchView')" :title="$t('buttons.switchView')" class="action" @click="change">
    <i class="material-icons">{{ icon }}</i>
    <span>{{ $t('buttons.switchView') }}</span>
  </button>
</template>

<script>
import { mapState, mapMutations } from 'vuex'
import { users as api } from '@/api'

export default {
  name: 'SwitchButton',
  computed: {
    ...mapState(['user']),
    icon: function() {
      if (this.user.viewMode === 'mosaic') return 'view_list'
      return 'view_module'
    }
  },
  methods: {
    ...mapMutations(['updateUser', 'closeHovers']),
    change: async function() {
      this.closeHovers()

      const data = {
        id: this.user.id,
        viewMode: (this.icon === 'view_list') ? 'list' : 'mosaic'
      }

      try {
        await api.update(data, ['viewMode'])
        this.updateUser(data)
      } catch (e) {
        this.$showError(e)
      }
    }
  }
}
</script>
