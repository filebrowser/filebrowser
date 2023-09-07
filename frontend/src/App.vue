<template>
  <div>
    <router-view></router-view>
  </div>
</template>

<script lang="ts">
export default {
  name: "app",
  mounted() {
    const loading = document.getElementById("loading");
    if(loading !== null) {
      loading.classList.add("done");

      setTimeout(function () {
        if(loading.parentNode !== null) {
          loading.parentNode.removeChild(loading);
        }
      }, 200);
    } 
  },
};
</script>