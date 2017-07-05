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
    CodeMirror.modeURL = this.$store.state.baseURL + '/static/js/codemirror/mode/%N/%N.js'

    this.content = CodeMirror.fromTextArea(document.getElementById('content'), {
      lineNumbers: (this.req.language !== 'markdown'),
      viewportMargin: Infinity
    })

    this.metadata = CodeMirror.fromTextArea(document.getElementById('metadata'), {
      viewportMargin: Infinity
    })

    CodeMirror.autoLoadMode(this.content, this.req.language)
  },
  methods: {
  }
}
</script>

<style>

</style>
