<template>
  <div class="prompt error">
    <i class="material-icons">error_outline</i>
    <h3>{{ $t('prompts.error') }}</h3>
    <pre>{{ $store.state.showMessage }}</pre>
    <div>
      <button @click="close" autofocus>{{ $t('buttons.close') }}</button>
      <button @click="reportIssue" class="cancel">{{ $t('buttons.reportIssue') }}</button>
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
