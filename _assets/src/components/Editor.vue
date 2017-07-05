<template>
    <form id="editor">
        <h2 v-if="hasMetadata">Metadata</h2>
        <textarea v-if="hasMetadata" id="metadata">{{ req.metadata }}</textarea>

        <h2 v-if="hasMetadata">Body</h2>
        <textarea id="content">{{ req.content }}</textarea>
    </form>
</template>

<script>
import { mapState } from 'vuex'
import CodeMirror from '@/codemirror'

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
      autofocus: true
    })

    CodeMirror.autoLoadMode(this.content, this.req.language)

    // Prevent of going on if there is no metadata.
    if (!this.hasMetadata) {
      return
    }

    this.metadata = CodeMirror.fromTextArea(document.getElementById('metadata'), {
      viewportMargin: Infinity
    })
  },
  methods: {
  }
}
</script>

<style>

</style>
