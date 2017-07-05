<template>
  <div class="prompt error">
    <i class="material-icons">error_outline</i>
    <h3>Something went wrong</h3>
    <pre>{{ error }}</pre>
    <div>
      <button @click="$store.commit('closeHovers')" autofocus>Close</button>
      <button @click="reportIssue" class="cancel">Report Issue</button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'error',
  props: ['error'],
  methods: {
    reportIssue () {
      window.open('https://github.com/hacdias/filemanager/issues/new')
    }
  }
}
</script>
