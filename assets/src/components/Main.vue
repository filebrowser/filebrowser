<template>
  <div>
    <site-header></site-header>
    <sidebar></sidebar>
    <main>
      <router-view></router-view>
    </main>
    <prompts></prompts>
  </div>
</template>

<script>
import Search from './Search'
import Sidebar from './Sidebar'
import Prompts from './prompts/Prompts'
import SiteHeader from './Header'

export default {
  name: 'main',
  components: {
    Search,
    Sidebar,
    SiteHeader,
    Prompts
  },
  watch: {
    '$route': function () {
      // Reset selected items and multiple selection.
      this.$store.commit('resetSelected')
      this.$store.commit('multiple', false)
      this.$store.commit('closeHovers')
    }
  }
}
</script>
