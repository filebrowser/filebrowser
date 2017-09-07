<template>
  <router-view @update:css="updateCSS" @clean:css="cleanCSS"></router-view>
</template>

<script>
export default {
  name: 'app',
  mounted () {
    // Remove loading animation.
    let loading = document.getElementById('loading')
    loading.classList.add('done')

    setTimeout(function () {
      loading.parentNode.removeChild(loading)
    }, 200)

    this.updateCSS()
  },
  methods: {
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
