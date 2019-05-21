<template>
  <div>
    <div id="progress">
      <div v-bind:style="{ width: $store.state.progress + '%' }"></div>
    </div>
    <site-header></site-header>
    <sidebar></sidebar>
    <main>
      <router-view></router-view>
      <shell v-if="isLogged && user.perm.execute" />
    </main>
    <prompts></prompts>
  </div>
</template>

<script>
import { mapState, mapGetters } from 'vuex'
import Sidebar from '@/components/Sidebar'
import Prompts from '@/components/prompts/Prompts'
import SiteHeader from '@/components/Header'
import Shell from '@/components/Shell'
import { context as api } from '@/api'

export default {
  name: 'layout',
  components: {
    Sidebar,
    SiteHeader,
    Prompts,
    Shell
  },
  computed: {
    ...mapGetters([ 'isLogged' ]),
    ...mapState([ 'user' ])
  },
  watch: {
    '$route': function () {
      this.$store.commit('resetSelected')
      this.$store.commit('multiple', false)
      if (this.$store.state.show !== 'success') this.$store.commit('closeHovers')
    }
  },
  async created() {
    try {
      const res = await api.get()
      this.$store.commit('updateContext', res)
    } catch (e) {
      this.$showError(e)
    }
  }
}
</script>
