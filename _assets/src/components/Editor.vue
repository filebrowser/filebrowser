<template>
    <form id="editor">
        <h2 v-if="editor.type == 'complete'">Metadata</h2>
    </form>
</template>

<script>
export default {
  name: 'editor',
  data: function () {
    return window.info.req.data
  },
  methods: {
  }
}
</script>
