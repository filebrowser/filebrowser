<template>
  <select name="selectLanguage" v-on:change="change" :value="locale">
    <option v-for="(language, value) in locales" :key="value" :value="value">
      {{ language }}
    </option>
  </select>
</template>

<script setup lang="ts">
import { markRaw } from "vue";

defineProps<{
  locale: string;
}>();

const emit = defineEmits<{
  "update:locale": [locale: string];
}>();

const locales = markRaw({
  he: "עברית",
  hr: "Hrvatski",
  hu: "Magyar",
  ar: "العربية",
  ca: "Català",
  cs: "Čeština",
  de: "Deutsch",
  el: "Ελληνικά",
  en: "English",
  es: "Español",
  fr: "Français",
  is: "Icelandic",
  it: "Italiano",
  ja: "日本語",
  ko: "한국어",
  "nl-be": "Dutch (Belgium)",
  no: "Norsk",
  pl: "Polski",
  "pt-br": "Português",
  pt: "Português (Brasil)",
  ro: "Romanian",
  ru: "Русский",
  sk: "Slovenčina",
  "sv-se": "Swedish (Sweden)",
  tr: "Türkçe",
  uk: "Українська",
  vi: "Tiếng Việt",
  "zh-cn": "中文 (简体)",
  "zh-tw": "中文 (繁體)",
});

const change = (event: Event) => {
  emit("update:locale", (event.target as HTMLSelectElement).value);
};
</script>
