<template>
  <select v-on:change="change" :value="locale">
    <option v-for="(language, value) in locales" :key="value" :value="value">{{ $t('languages.' + language) }}</option>
  </select>
</template>

<script>

export default {
  name: 'languages',
  props: [ 'locale' ],
  data() {
    let dataObj = {
      locales: {
        ar: 'ar',
        de: 'de',
        en: 'en',
        es: 'es',
        fr: 'fr',
        is: 'is',
        it: 'it',
        ja: 'ja',
        ko: 'ko',
        'nl-be': 'nlBE',
        pl: 'pl',
        'pt-br': 'ptBR',
        pt: 'pt',
        ro: 'ro',
        ru: 'ru',
        'sv-se': 'svSE',
        'zh-cn': 'zhCN',
        'zh-tw': 'zhTW'
      }
    };

    Object.defineProperty(dataObj, "locales", { configurable: false, writable: false });

    return dataObj;
  },
  methods: {
    change (event) {
      this.$emit('update:locale', event.target.value)
    }
  }
}
</script>
