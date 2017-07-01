<template>
  <div class="prompt">
    <h3>Rename</h3>
    <p>Insert a new name for <code>{{ oldName() }}</code>:</p>
    <input autofocus type="text" @keyup.enter="submit" v-model.trim="name">
    <div>
      <button @click="submit" type="submit">Rename</button>
      <button @click="cancel" class="cancel">Cancel</button>
    </div>
  </div>
</template>

<script>
import { mapState } from 'vuex'
import page from '../utils/page'
import webdav from '../utils/webdav'

export default {
  name: 'rename-prompt',
  data: function () {
    return {
      name: ''
    }
  },
  computed: mapState(['req', 'selected', 'selectedCount']),
  methods: {
    cancel: function (event) {
      this.$store.commit('showRename', false)
    },
    oldName: function () {
      if (this.req.kind !== 'listing') {
        return this.req.data.name
      }

      if (this.selectedCount === 0 || this.selectedCount > 1) {
        // This shouldn't happen.
        return
      }

      return this.req.data.items[this.selected[0]].name
    },
    submit: function (event) {
      let oldLink = ''
      let newLink = ''

      if (this.req.kind !== 'listing') {
        oldLink = this.req.data.url
      } else {
        oldLink = this.req.data.items[this.selected[0]].url
      }

      this.name = encodeURIComponent(this.name)
      newLink = page.removeLastDir(oldLink) + '/' + this.name

      // buttons.setLoading('rename')

      webdav.move(oldLink, newLink)
        .then(() => {
          if (this.req.kind !== 'listing') {
            page.open(newLink)
            return
          }
          // TODO: keep selected after reload?
          page.reload()
          // buttons.setDone('rename')
        }).catch(error => {
          // buttons.setDone('rename', false)
          console.log(error)
        })

      this.$store.commit('showRename', false)
      return
    }
  }
}
</script>
