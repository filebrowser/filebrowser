<template>
  <button @click="change" :aria-label="$t('buttons.switchView')" :title="$t('buttons.switchView')" class="action" id="switch-view-button">
    <i class="material-icons">{{ icon }}</i>
    <span>{{ $t('buttons.switchView') }}</span>
  </button>
</template>

<script>
import { mapState, mapMutations } from 'vuex'
import { updateUser } from '@/utils/api'

export default {
  name: 'switch-button',
  computed: {
    ...mapState(['user']),
    icon: function () {
      if (this.user.viewMode === 'mosaic') return 'view_list'
      return 'view_module'
    }
  },
  methods: {
    ...mapMutations(['updateUser']),
    change: function (event) {
      // If we are on mobile we should close the dropdown.
      this.$store.commit('closeHovers')

      let user = {...this.user}
      user.viewMode = (this.icon === 'view_list') ? 'list' : 'mosaic'

      updateUser(user, 'partial').then(() => {
        this.updateUser({ viewMode: user.viewMode })
      }).catch(this.$showError)
    }
  }
}
</script>
