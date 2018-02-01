<template>
    <form id="editor" :class="req.language">
        <div v-if="hasMetadata" id="metadata">
          <h2>{{ $t('files.metadata') }}</h2>
        </div>

        <h2 v-if="hasMetadata">{{ $t('files.body') }}</h2>
    </form>
</template>

<script>
import { mapState } from 'vuex'
import CodeMirror from '@/utils/codemirror'
import * as api from '@/utils/api'
import buttons from '@/utils/buttons'

export default {
  name: 'editor',
  computed: {
    ...mapState(['req', 'schedule']),
    hasMetadata: function () {
      return (this.req.metadata !== undefined && this.req.metadata !== null)
    }
  },
  data: function () {
    return {
      metadata: null,
      metalang: null,
      content: null
    }
  },
  created () {
    window.addEventListener('keydown', this.keyEvent)
    document.getElementById('save-button').addEventListener('click', this.save)

    let publish = document.getElementById('publish-button')
    if (publish !== null) {
      publish.addEventListener('click', this.publish)
    }
  },
  beforeDestroy () {
    window.removeEventListener('keydown', this.keyEvent)
    document.getElementById('save-button').removeEventListener('click', this.save)

    let publish = document.getElementById('publish-button')
    if (publish !== null) {
      publish.removeEventListener('click', this.publish)
    }
  },
  mounted: function () {
    if (this.req.content === undefined || this.req.content === null) {
      this.req.content = ''
    }

    // Set up the main content editor.
    this.content = CodeMirror(document.getElementById('editor'), {
      value: this.req.content,
      lineNumbers: (this.req.language !== 'markdown'),
      viewportMargin: 500,
      autofocus: true,
      mode: this.req.language,
      theme: (this.req.language === 'markdown') ? 'markdown' : 'ttcn',
      lineWrapping: (this.req.language === 'markdown')
    })

    CodeMirror.autoLoadMode(this.content, this.req.language)

    // Prevent of going on if there is no metadata.
    if (!this.hasMetadata) {
      return
    }

    this.parseMetadata()

    // Set up metadata editor.
    this.metadata = CodeMirror(document.getElementById('metadata'), {
      value: this.req.metadata,
      viewportMargin: Infinity,
      lineWrapping: true,
      theme: 'markdown',
      mode: this.metalang
    })

    CodeMirror.autoLoadMode(this.metadata, this.metalang)
  },
  methods: {
    // Saves the content when the user presses CTRL-S.
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
    // Parses the metadata and gets the language in which
    // it is written.
    parseMetadata () {
      if (this.req.metadata.startsWith('{')) {
        this.metalang = 'json'
      }

      if (this.req.metadata.startsWith('---')) {
        this.metalang = 'yaml'
      }

      if (this.req.metadata.startsWith('+++')) {
        this.metalang = 'toml'
      }
    },
    // Publishes the file.
    publish (event) {
      this.save(event, true)
    },
    // Saves the file.
    save (event, regenerate = false) {
      let button = regenerate ? 'publish' : 'save'
      if (this.schedule !== '') button = 'schedule'
      let content = this.content.getValue()
      buttons.loading(button)

      if (this.hasMetadata) {
        content = this.metadata.getValue() + '\n\n' + content
      }

      api.put(this.$route.path, content, regenerate, this.schedule)
        .then(() => {
          buttons.success(button)
          this.$store.commit('setSchedule', '')
        })
        .catch(error => {
          buttons.done(button)
          this.$showError(error)
          this.$store.commit('setSchedule', '')
        })
    }
  }
}
</script>
