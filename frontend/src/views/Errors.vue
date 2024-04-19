<template>
  <div>
    <header-bar v-if="showHeader" showMenu showLogo />

    <h2 class="message">
      <i class="material-icons">{{ info.icon }}</i>
      <span>{{ t(info.message) }}</span>
    </h2>
  </div>
</template>

<script setup lang="ts">
import HeaderBar from "@/components/header/HeaderBar.vue";
import { computed } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n({});

const errors: {
  [key: number]: {
    icon: string;
    message: string;
  };
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

const props = withDefaults(
  defineProps<{
    errorCode?: number;
    showHeader?: boolean;
  }>(),
  {
    errorCode: 500,
    showHeader: false,
  }
);

const info = computed(() => {
  return errors[props.errorCode] ? errors[props.errorCode] : errors[500];
});
</script>
