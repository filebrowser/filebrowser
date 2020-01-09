<template>
  <form id="editor"></form>
</template>

<script>
import { mapState } from 'vuex'
import { files as api } from '@/api'
import buttons from '@/utils/buttons'

import ace from 'ace-builds/src-min-noconflict/ace.js'
import modelist from 'ace-builds/src-min-noconflict/ext-modelist.js'
import 'ace-builds/webpack-resolver'
import { theme } from '@/utils/constants'

export default {
  name: 'editor',
  computed: {
    ...mapState(['req'])
  },
  data: function () {
    return {}
  },
  created () {
    window.addEventListener('keydown', this.keyEvent)
    document.getElementById('save-button').addEventListener('click', this.save)
  },
  beforeDestroy () {
    window.removeEventListener('keydown', this.keyEvent)
    document.getElementById('save-button').removeEventListener('click', this.save)
    this.editor.destroy();
  },
  mounted: function () {
    const fileContent = this.req.content || '';

    this.editor = ace.edit('editor', {
      maxLines: Infinity,
      minLines: 20,
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
