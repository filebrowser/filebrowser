<template>
  <div class="prompt success">
    <i class="material-icons">done</i>
    <h3>{{ $store.state.showMessage }}</h3>
    <div>
      <button @click="close" autofocus>OK</button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'success',
  methods: {
    close () {
      this.$store.commit('closeHovers')
    }
  }
}
</script>
