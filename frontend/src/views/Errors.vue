<template>
  <div>
    <header-bar v-if="showHeader" showMenu showLogo />

    <h2 class="message">
      <i class="material-icons">{{ info.icon }}</i>
      <span>{{ $t(info.message) }}</span>
    </h2>
  </div>
</template>

<script>
import HeaderBar from "@/components/header/HeaderBar";

const errors = {
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

export default {
  name: "errors",
  components: {
    HeaderBar,
  },
  props: ["errorCode", "showHeader"],
  computed: {
    code() {
      return this.errorCode === "0" ||
        this.errorCode === "404" ||
        this.errorCode === "403"
        ? parseInt(this.errorCode)
        : 500;
    },
    info() {
      return errors[this.code];
    },
  },
};
</script>
