<template>
  <select name="selectLanguage" v-on:change="change" :value="locale">
    <option v-for="(language, value) in locales" :key="value" :value="value">
      {{ language }}
    </option>
  </select>
</template>

<script>
import { markRaw } from "vue";

export default {
  name: "languages",
  props: ["locale"],
  data() {
    let dataObj = {};
    const locales = {
      ar_AR: "العربية",
      en_GB: "English",
      es_ES: "Español",
      es_AR: "Español (Argentina)",
      es_CO: "Español (Colombia)",
      es_MX: "Español (Mexico)",
      fr_FR: "Français",
      id_ID: "Bahasa Indonesia",
      lt_LT: "Lietuvių",
      pt_BR: "Português (Brasil)",
      pt_PT: "Português",
      ru_RU: "Русский",
      tr_TR: "Türkçe",
      uk_UA: "Український",
      zh_CN: "中文 (简体)"
    };

    // Vue3 reactivity breaks with this configuration
    // so we need to use markRaw as a workaround
    // https://github.com/vuejs/core/issues/3024
    Object.defineProperty(dataObj, "locales", {
      value: markRaw(locales),
      configurable: false,
      writable: false,
    });

    return dataObj;
  },
  methods: {
    change(event) {
      this.$emit("update:locale", event.target.value);
    },
  },
};
</script>
