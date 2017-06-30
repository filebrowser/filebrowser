<template>
  <div class="prompt">
    <h3>Delete files</h3>
    <p v-show="req.kind !== 'listing'">Are you sure you want to delete this file/folder?</p>
    <p v-show="req.kind === 'listing'">Are you sure you want to delete {{ selected.length }} file(s)?</p>
    <div>
      <button @click="submit" autofocus>Delete</button>
      <button @click="cancel" class="cancel">Cancel</button>
    </div>
  </div>
</template>

<script>
import webdav from '../webdav'
import page from '../page'

var $ = window.info

export default {
  name: 'delete-prompt',
  data: function () {
    return window.info
  },
  methods: {
    cancel: function (event) {
      this.$store.commit('showDelete', false)
    },
    submit: function (event) {
      this.$store.commit('showDelete', false)
      // buttons.setLoading('delete')

      if ($.req.kind !== 'listing') {
        webdav.trash(window.location.pathname)
          .then(() => {
            // buttons.setDone('delete')
            page.open(page.removeLastDir(window.location.pathname))
          })
          .catch(error => {
            // buttons.setDone('delete', false)
            console.log(error)
          })

        return
      }

      if ($.selected.length === 0) {
        // This shouldn't happen...
        return
      }

      if ($.selected.length === 1) {
        webdav.trash($.req.data.items[$.selected[0]].url)
          .then(() => {
            // buttons.setDone('delete')
            page.reload()
          })
          .catch(error => {
            // buttons.setDone('delete', false)
            console.log(error)
          })

        return
      }

      // More than one item!
      let promises = []

      for (let index of $.selected) {
        promises.push(webdav.trash($.req.data.items[index].url))
      }

      Promise.all(promises)
        .then(() => {
          page.reload()
          // buttons.setDone('delete')
        })
        .catch(error => {
          console.log(error)
          // buttons.setDone('delete', false)
        })
    }
  }
}
</script>
