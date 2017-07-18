<template>
  <div class="prompt error">
    <i class="material-icons">error_outline</i>
    <h3>Something went wrong</h3>
    <pre>{{ $store.state.showMessage }}</pre>
    <div>
      <button @click="close" autofocus>Close</button>
      <button @click="reportIssue" class="cancel">Report Issue</button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'error',
  methods: {
    reportIssue () {
      window.open('https://github.com/hacdias/filemanager/issues/new')
    },
    close () {
      this.$store.commit('closeHovers')
    }
  }
}
</script>
