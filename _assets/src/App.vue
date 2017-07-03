<template>
  <router-view></router-view>
</template>

<script>
export default {
  name: 'app',
  mounted: function () {
    let loading = document.getElementById('loading')
    loading.classList.add('done')

    setTimeout(function () {
      loading.parentNode.removeChild(loading)
    }, 200)
  }
}
</script>

<style>
@import './css/styles.css';
</style>
