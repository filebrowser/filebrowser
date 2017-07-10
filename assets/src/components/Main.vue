<template>
  <div>
    <site-header></site-header>
    <sidebar></sidebar>
    <main>
      <router-view v-on:css-updated="updateCSS"></router-view>
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
      this.$store.commit('resetSelected')
      this.$store.commit('multiple', false)
      if (this.$store.state.show !== 'success') this.$store.commit('closeHovers')
    }
  },
  mounted () {
    this.updateCSS()
  },
  methods: {
    updateCSS () {
      let css = this.$store.state.user.css

      let style = document.querySelector('style[title="user-css"]')
      if (style !== undefined && style !== null) {
        style.parentElement.removeChild(style)
      }

      style = document.createElement('style')
      style.title = 'user-css'
      style.type = 'text/css'
      style.appendChild(document.createTextNode(css))
      document.head.appendChild(style)
    }
  }
}
</script>
