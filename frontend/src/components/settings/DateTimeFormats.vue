<template>
  <select
    name="selectDateTimeFormat"
    v-on:change="change"
    :value="dateTimeFormat"
  >
    <option
      v-for="(dateTimeFormat, value) in dateTimeFormats"
      :key="value"
      :value="value"
    >
      {{ dateTimeFormat }}
    </option>
  </select>
</template>

<script>
import { markRaw } from "vue";

export default {
  name: "dateTimeFormats",
  props: ["dateTimeFormat"],
  data() {
    let dataObj = {};
    const dateTimeFormats = {
      "MM/DD/YYYY h:mm A": "02/21/2023 3:59 PM",
      "YYYY/MM/DD HH:mm": "2023/02/21 15:59",
      "DD/MM/YYYY HH:mm": "21/02/2023 15:59",
    };

    // Vue3 reactivity breaks with this configuration
    // so we need to use markRaw as a workaround
    // https://github.com/vuejs/core/issues/3024
    Object.defineProperty(dataObj, "dateTimeFormats", {
      value: markRaw(dateTimeFormats),
      configurable: false,
      writable: false,
    });

    return dataObj;
  },
  methods: {
    change(event) {
      this.$emit("update:dateTimeFormat", event.target.value);
    },
  },
};
</script>
