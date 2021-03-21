<template>
  <div>
    <header-bar v-if="showHeader" showMenu showLogo />

    <h2 class="message">
      <i class="material-icons">{{ icon }}</i>
      <span>{{ message }}</span>
    </h2>
  </div>
</template>

<script>
import HeaderBar from "@/components/header/HeaderBar";

const errors = {
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

export default {
  name: "errors",
  components: {
    HeaderBar,
  },
  props: ["errorCode", "showHeader"],
  data: function () {
    return {
      icon: errors[this.errorCode].icon,
      message: this.$t(errors[this.errorCode].message),
    };
  },
};
</script>
