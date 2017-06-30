<template>
  <div class="prompt">
    <h3>New file</h3>
    <p>Write the name of the new file.</p>
    <input autofocus type="text" @keyup.enter="submit" v-model.trim="name">
    <div>
      <button class="ok" @click="submit">Create</button>
      <button class="cancel" @click="$store.commit('showNewFile', false)">Cancel</button>
    </div>
  </div>
</template>

<script>
import page from '../utils/page'
import webdav from '../utils/webdav'

export default {
  name: 'new-file-prompt',
  data: function () {
    return {
      name: ''
    }
  },
  methods: {
    submit: function (event) {
      event.preventDefault()
      if (this.new === '') return

      // buttons.setLoading('newFile')
      webdav.create(window.location.pathname + this.name)
        .then(() => {
          // buttons.setDone('newFile')
          page.open(window.location.pathname + this.name)
        })
        .catch(e => {
          // buttons.setDone('newFile', false)
          console.log(e)
        })

      this.$store.commit('showNewFile', false)
    }
  }
}
</script>

