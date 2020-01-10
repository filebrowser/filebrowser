<template>
  <div id="previewer">
    <div class="bar">
      <button @click="back" class="action" :title="$t('files.closePreview')" :aria-label="$t('files.closePreview')" id="close">
        <i class="material-icons">close</i>
      </button>
      <span class="title">{{ req.name + ((contentChanged) ? '*' : '') }}</span>
      <button @click="save" v-show="user.perm.modify" :aria-label="$t('buttons.save')" :title="$t('buttons.save')" class="action" id="save-button">
        <i class="material-icons">save</i>
      </button>
    </div>
    <div class="editor">
      <form id="editor"></form>
    </div>
  </div>
</template>

<script>
import { mapState } from 'vuex'
import { files as api } from '@/api'
import buttons from '@/utils/buttons'

import ace from 'ace-builds/src-min-noconflict/ace.js'
import modelist from 'ace-builds/src-min-noconflict/ext-modelist.js'
import 'ace-builds/webpack-resolver'

export default {
  name: 'editor',
  computed: {
    ...mapState(['req', 'user', 'previewContent'])
  },
  data: function () {
    return {
      editor: null,
      contentChanged: false
    }
  },
  created () {
    window.addEventListener('keydown', this.keyEvent)
  },
  beforeDestroy () {
    window.removeEventListener('keydown', this.keyEvent)
  },
  mounted: function () {
    this.editor = ace.edit('editor', {
      maxLines: Infinity,
      minLines: 20,
      value: this.previewContent,
      showPrintMargin: false,
      readOnly: this.req.type === 'textImmutable',
      theme: 'ace/theme/twilight',
      mode: modelist.getModeForPath(this.req.name).mode,
      wrap: true
    })

    this.editor.on('change', () => {
      this.contentChanged = true
    })
  },
  methods: {
    keyEvent (event) {
      let key = String.fromCharCode(event.which).toLowerCase()

      if ((event.ctrlKey || event.metaKey) && key == 's') {
          event.preventDefault()
          this.save()
      } else if (event.which === 27) { // Esc
          this.$store.commit('toggleEditor')
      }
    },
    async save () {
      const button = 'save'
      buttons.loading('save')

      try {
        await api.put(this.$route.path, this.editor.getValue())

        this.$store.commit('setPreviewContent', this.editor.getValue())
        this.contentChanged = false

        buttons.success(button)
      } catch (e) {
        buttons.done(button)
        this.$showError(e)
      }
    },
    back () {
      this.$store.commit('toggleEditor')
    }
  }
}
</script>
