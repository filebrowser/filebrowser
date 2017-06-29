<template>
  <div class="prompt">
    <h3>New file</h3>
    <p>Write the name of the new file.</p>
    <input autofocus type="text" @keyup.enter="submit" v-model.trim="name">
    <div>
      <button class="ok" @click="submit">Create</button>
      <button class="cancel" @click="cancel">Cancel</button>
    </div>
  </div>
</template>

<script>
import page from '../page'
import webdav from '../webdav'

var $ = window.info

export default {
  name: 'new-file-prompt',
  data: function () {
    return {
      name: ''
    }
  },
  methods: {
    cancel: function () {
      $.showNewFile = false
    },
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

      $.showNewFile = false
    }
  }
}
</script>

