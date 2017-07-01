<template>
    <form id="editor">
        <h2 v-if="req.data.editor.type == 'complete'">Metadata</h2>

        <h2 v-if="req.data.editor.type == 'complete'">Body</h2>

        <div v-if="req.data.editor.type !== 'frontmatter-only'" class="content">
            <div id="ace"></div>
            <textarea id="source" name="content">{{ req.data.content }}</textarea>
        </div>
    </form>
</template>

<script>
import { mapState } from 'vuex'

export default {
  name: 'editor',
  computed: mapState(['req']),
  data: function () {
    return {
      codemirror: null,
      simplemde: null
    }
  },
  mounted: function () {

  },
  methods: {
  }
}
</script>
