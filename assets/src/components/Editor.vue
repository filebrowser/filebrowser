<template>
    <form id="editor" :class="req.language">
        <h2 v-if="hasMetadata">Metadata</h2>
        <textarea v-model="req.metadata" v-if="hasMetadata" id="metadata"></textarea>

        <h2 v-if="hasMetadata">Body</h2>
        <textarea v-model="req.content" id="content"></textarea>
    </form>
</template>

<script>
import { mapState } from 'vuex'
import CodeMirror from '@/utils/codemirror'

export default {
  name: 'editor',
  computed: {
    ...mapState(['req']),
    hasMetadata: function () {
      return (this.req.metadata !== undefined && this.req.metadata !== null)
    }
  },
  data: function () {
    return {
      metadata: null,
      content: null
    }
  },
  mounted: function () {
    this.content = CodeMirror.fromTextArea(document.getElementById('content'), {
      lineNumbers: (this.req.language !== 'markdown'),
      viewportMargin: Infinity,
      autofocus: true,
      theme: (this.req.language === 'markdown') ? 'markdown' : 'ttcn',
      lineWrapping: (this.req.language === 'markdown')
    })

    CodeMirror.autoLoadMode(this.content, this.req.language)

    // Prevent of going on if there is no metadata.
    if (!this.hasMetadata) {
      return
    }

    this.metadata = CodeMirror.fromTextArea(document.getElementById('metadata'), {
      viewportMargin: Infinity,
      lineWrapping: true,
      theme: 'markdown'
    })

    if (this.req.metadata.startsWith('{')) {
      CodeMirror.autoLoadMode(this.metadata, 'json')
    }

    if (this.req.metadata.startsWith('---')) {
      CodeMirror.autoLoadMode(this.metadata, 'yaml')
    }

    if (this.req.metadata.startsWith('+++')) {
      CodeMirror.autoLoadMode(this.metadata, 'toml')
    }
  },
  methods: {
  }
}
</script>

<style>

</style>
