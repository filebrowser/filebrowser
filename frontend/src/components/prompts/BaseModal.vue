<template>
  <VueFinalModal
    class="vfm-modal"
    overlay-transition="vfm-fade"
    content-transition="vfm-fade"
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
</script>
