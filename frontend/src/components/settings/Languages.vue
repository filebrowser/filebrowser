<template>
  <select v-on:change="change" :value="locale">
    <option value="ar">{{ $t('languages.ar') }}</option>
    <option value="de">{{ $t('languages.de') }}</option>
    <option value="en">{{ $t('languages.en') }}</option>
    <option value="es">{{ $t('languages.es') }}</option>
    <option value="fr">{{ $t('languages.fr') }}</option>
    <option value="it">{{ $t('languages.it') }}</option>
    <option value="ja">{{ $t('languages.ja') }}</option>
    <option value="ko">{{ $t('languages.ko') }}</option>
    <option value="pl">{{ $t('languages.pl') }}</option>
    <option value="pt-br">{{ $t('languages.ptBR') }}</option>
    <option value="pt">{{ $t('languages.pt') }}</option>
    <option value="ru">{{ $t('languages.ru') }}</option>
    <option value="zh-cn">{{ $t('languages.zhCN') }}</option>
    <option value="zh-tw">{{ $t('languages.zhTW') }}</option>
  </select>
</template>

<script>
export default {
  name: 'languages',
  props: [ 'locale' ],
  methods: {
    change (event) {
      this.$emit('update:locale', event.target.value)
    }
  }
}
</script>
