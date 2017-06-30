<template>
  <div class="prompt">
    <h3>New directory</h3>
    <p>Write the name of the new directory.</p>
    <input autofocus type="text" @keyup.enter="submit" v-model.trim="name">
    <div>
      <button class="ok" @click="submit">Create</button>
      <button class="cancel" @click="$store.commit('showNewDir', false)">Cancel</button>
    </div>
  </div>
</template>

<script>
import page from '../utils/page'
import webdav from '../utils/webdav'

export default {
  name: 'new-dir-prompt',
  data: function () {
    return {
      name: ''
    }
  },
  methods: {
    submit: function (event) {
      event.preventDefault()
      if (this.new === '') return

      let url = window.location.pathname + this.name + '/'
      url = url.replace('//', '/')

      // buttons.setLoading('newDir')
      webdav.create(url)
        .then(() => {
          // buttons.setDone('newDir')
          page.open(url)
        })
        .catch(e => {
          // buttons.setDone('newDir', false)
          console.log(e)
        })

      this.$store.commit('showNewDir', false)
    }
  }
}
</script>

