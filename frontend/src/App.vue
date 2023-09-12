<template>
  <div>
    <router-view></router-view>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from "vue";

onMounted(() => {
  // we can safely assume non-nullity here
  const loading = document.getElementById("loading")!;
  loading.classList.add("done");

  setTimeout(function () {
    loading.parentNode!.removeChild(loading);
  }, 200);
});
</script>
