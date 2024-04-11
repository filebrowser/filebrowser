<template>
  <select name="selectLanguage" v-on:change="change" :value="locale">
    <option v-for="(language, value) in locales" :key="value" :value="value">
      {{ $t("languages." + language) }}
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
      he: "he",
      hu: "hu",
      ar: "ar",
      de: "de",
      el: "el",
      en: "en",
      es: "es",
      fr: "fr",
      is: "is",
      it: "it",
      ja: "ja",
      ko: "ko",
      "nl-be": "nlBE",
      pl: "pl",
      "pt-br": "ptBR",
      pt: "pt",
      ro: "ro",
      ru: "ru",
      sk: "sk",
      "sv-se": "svSE",
      tr: "tr",
      uk: "uk",
      "zh-cn": "zhCN",
      "zh-tw": "zhTW",
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
