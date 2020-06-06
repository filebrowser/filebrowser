<template>
  <select :value="theme" @change="change">
    <option value="">{{ $t('settings.themes.light') }}</option>
    <option value="dark">{{ $t('settings.themes.dark') }}</option>
  </select>
</template>

<script>
export default {
  name: 'Themes',
  props: ['theme'],
  methods: {
    change(event) {
      this.$emit('update:theme', event.target.value)
    }
  }
}
</script>
