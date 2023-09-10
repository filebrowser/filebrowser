<template>
  <div>
    <router-view></router-view>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from "vue";

onMounted(() => {
  const loading = document.getElementById("loading");
  if (loading !== null) {
    loading.classList.add("done");

    setTimeout(function () {
      if (loading.parentNode !== null) {
        loading.parentNode.removeChild(loading);
      }
    }, 200);
  }
});
</script>
