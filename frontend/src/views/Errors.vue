<template>
  <div>
    <header-bar v-if="showHeader" showMenu showLogo />

    <h2 class="message">
      <i class="material-icons">{{ info.icon }}</i>
      <span>{{ $t(info.message) }}</span>
    </h2>
  </div>
</template>

<script setup lang="ts">
import HeaderBar from "@/components/header/HeaderBar.vue";
import { computed } from "vue";

const errors: {
  [key: string]: {
    icon: string,
    message: string
  }
} = {
  0: {
    icon: "cloud_off",
    message: "errors.connection",
  },
  403: {
    icon: "error",
    message: "errors.forbidden",
  },
  404: {
    icon: "gps_off",
    message: "errors.notFound",
  },
  500: {
    icon: "error_outline",
    message: "errors.internal",
  },
};

const props = defineProps<{
  errorCode: any,
  showHeader: boolean
}>()

const info = computed(() => {
  return errors[props.errorCode] ? errors [props.errorCode] : errors[500]
})

</script>
