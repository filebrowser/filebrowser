<template>
  <VueFinalModal
    class="vfm-modal"
    overlay-transition="vfm-fade"
    content-transition="vfm-fade"
    @closed="layoutStore.closeHovers"
    :focus-trap="{
      initialFocus: '#focus-prompt',
      fallbackFocus: 'div.vfm__content',
    }"
  >
    <slot />
  </VueFinalModal>
</template>

<script setup lang="ts">
import { VueFinalModal } from "vue-final-modal";
import { useLayoutStore } from "@/stores/layout";

const layoutStore = useLayoutStore();
</script>
