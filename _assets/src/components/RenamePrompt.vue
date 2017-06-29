<template>
  <div class="prompt">
    <h3>Rename</h3>
    <p>Insert a new name for <code>{{ oldName() }}</code>:</p>
    <input autofocus type="text" @keyup.enter="submit" v-model.trim="name">
    <div>
      <button @click="submit" type="submit" autofocus>Rename</button>
      <button @click="cancel" class="cancel">Cancel</button>
    </div>
  </div>
</template>

<script>
import page from '../page'
import webdav from '../webdav'

var $ = window.info

export default {
  name: 'rename-prompt',
  data: function () {
    return {
      name: ''
    }
  },
  methods: {
    cancel: function (event) {
      $.showRename = false
      this.name = ''
    },
    oldName: function () {
      if ($.req.kind !== 'listing') {
        return $.req.data.name
      }

      if ($.listing.selected.length === 0 || $.listing.selected.length > 1) {
        // This shouldn't happen.
        return
      }

      return $.req.data.items[$.listing.selected[0]].name
    },
    submit: function (event) {
      let oldLink = ''
      let newLink = ''

      if ($.req.kind !== 'listing') {
        oldLink = $.req.data.url
      } else {
        oldLink = $.req.data.items[$.listing.selected[0]].url
      }

      newLink = page.removeLastDir(oldLink) + '/' + this.name

      // buttons.setLoading('rename')

      webdav.move(oldLink, newLink)
        .then(() => {
          // TODO: keep selected after reload?
          page.reload()
          // buttons.setDone('rename')
        }).catch(error => {
          // buttons.setDone('rename', false)
          console.log(error)
        })

      this.name = ''
      $.showRename = false
      return
    }
  }
}
</script>
