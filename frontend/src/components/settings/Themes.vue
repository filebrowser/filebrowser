<template>
  <select v-on:change="change" :value="theme">
    <option value="">{{ $t('settings.themes.light') }}</option>
    <option value="dark">{{ $t('settings.themes.dark') }}</option>
  </select>
</template>

<script>
export default {
  name: 'themes',
  props: [ 'theme' ],
  methods: {
    change (event) {
      this.$emit('update:theme', event.target.value)
    }
  }
}
</script>