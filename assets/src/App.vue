<template>
  <router-view :dependencies="loaded" @update:css="updateCSS" @clean:css="cleanCSS"></router-view>
</template>

<script>
import { mapState } from 'vuex'

export default {
  name: 'app',
  computed: mapState(['recaptcha']),
  data () {
    return {
      loaded: false
    }
  },
  mounted () {
    if (this.recaptcha.length === 0) {
      this.unload()
      return
    }

    let check = () => {
      if (typeof window.grecaptcha === 'undefined') {
        setTimeout(check, 100)
        return
      }

      this.unload()
    }

    check()
  },
  methods: {
    unload () {
      this.loaded = true
      // Remove loading animation.
      let loading = document.getElementById('loading')
      loading.classList.add('done')

      setTimeout(function () {
        loading.parentNode.removeChild(loading)
      }, 200)

      this.updateCSS()
    },
    updateCSS (global = false) {
      let css = this.$store.state.css

      if (typeof this.$store.state.user.css === 'string' && !global) {
        css += '\n' + this.$store.state.user.css
      }

      this.removeCSS()

      let style = document.createElement('style')
      style.title = 'custom-css'
      style.type = 'text/css'
      style.appendChild(document.createTextNode(css))
      document.head.appendChild(style)
    },
    removeCSS () {
      let style = document.querySelector('style[title="custom-css"]')
      if (style === undefined || style === null) {
        return
      }

      style.parentElement.removeChild(style)
    },
    cleanCSS () {
      this.updateCSS(true)
    }
  }
}
</script>

<style>
@import './css/styles.css';
</style>
