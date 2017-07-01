<template>
    <form id="editor">
        <h2 v-if="hasMetadata">Metadata</h2>

        <h2 v-if="hasMetadata">Body</h2>

        <div class="content">
            <div id="ace"></div>
            <textarea id="source" name="content">{{ req.data.content }}</textarea>
        </div>
    </form>
</template>

<script>
import { mapState } from 'vuex'

export default {
  name: 'editor',
  computed: {
    ...mapState(['req']),
    hasMetadata: function () {
      return (this.req.data.metadata !== undefined && this.req.data.metadata !== null)
    }
  },
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
