<template>
  <div id="editor-container">
    <div class="bar">
      <button @click="back" :title="$t('files.closePreview')" :aria-label="$t('files.closePreview')" id="close" class="action">
        <i class="material-icons">close</i>
      </button>

      <div class="title">
        <span>{{ req.name }}</span>
      </div>

      <button @click="save" v-show="user.perm.modify" :aria-label="$t('buttons.save')" :title="$t('buttons.save')" id="save-button" class="action">
        <i class="material-icons">save</i>
      </button>
    </div>

    <div id="breadcrumbs">
      <span><i class="material-icons">home</i></span>

      <span v-for="(link, index) in breadcrumbs" :key="index">
        <span class="chevron"><i class="material-icons">keyboard_arrow_right</i></span>
        <span>{{ link.name }}</span>
      </span>
    </div>

    <form id="editor"></form>
  </div>
</template>

<script>
import { mapState } from 'vuex'
import { files as api } from '@/api'
import buttons from '@/utils/buttons'
import url from '@/utils/url'

import ace from 'ace-builds/src-min-noconflict/ace.js'
import modelist from 'ace-builds/src-min-noconflict/ext-modelist.js'
import 'ace-builds/webpack-resolver'
import { theme } from '@/utils/constants'

export default {
  name: 'editor',
  data: function () {
    return {}
  },
  computed: {
    ...mapState(['req', 'user']),
    breadcrumbs () {
      let parts = this.$route.path.split('/')

      if (parts[0] === '') {
        parts.shift()
      }

      if (parts[parts.length - 1] === '') {
        parts.pop()
      }

      let breadcrumbs = []

      for (let i = 0; i < parts.length; i++) {
        breadcrumbs.push({ name: decodeURIComponent(parts[i]) })
      }

      breadcrumbs.shift()

      if (breadcrumbs.length > 3) {
        while (breadcrumbs.length !== 4) {
          breadcrumbs.shift()
        }

        breadcrumbs[0].name = '...'
      }

      return breadcrumbs
    }
  },
  created () {
    window.addEventListener('keydown', this.keyEvent)
  },
  beforeDestroy () {
    window.removeEventListener('keydown', this.keyEvent)
    this.editor.destroy();
  },
  mounted: function () {    
    const fileContent = this.req.content || '';

    this.editor = ace.edit('editor', {
      value: fileContent,
      showPrintMargin: false,
      readOnly: this.req.type === 'textImmutable',
      theme: 'ace/theme/chrome',
      mode: modelist.getModeForPath(this.req.name).mode,
      wrap: true
    })

    if (theme == 'dark') {
      this.editor.setTheme("ace/theme/twilight");
    }
  },
  methods: {
    back () {
      let uri = url.removeLastDir(this.$route.path) + '/'
      this.$router.push({ path: uri })
    },
    keyEvent (event) {
      if (!event.ctrlKey && !event.metaKey) {
        return
      }

      if (String.fromCharCode(event.which).toLowerCase() !== 's') {
        return
      }

      event.preventDefault()
      this.save()
    },
    async save () {
      const button = 'save'
      buttons.loading('save')

      try {
        await api.put(this.$route.path, this.editor.getValue())
        buttons.success(button)
      } catch (e) {
        buttons.done(button)
        this.$showError(e)
      }
    }
  }
}
</script>
