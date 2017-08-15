<template>
  <div>
    <div id="progress">
      <div v-bind:style="{ width: $store.state.progress + '%' }"></div>
    </div>
    <site-header></site-header>
    <sidebar></sidebar>
    <main>
      <router-view v-on:css-updated="updateCSS"></router-view>
    </main>
    <prompts></prompts>
  </div>
</template>

<script>
import Search from '@/components/Search'
import Sidebar from '@/components/Sidebar'
import Prompts from '@/components/prompts/Prompts'
import SiteHeader from '@/components/Header'

export default {
  name: 'layout',
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
