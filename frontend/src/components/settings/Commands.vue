<template>
  <div>
    <h3>{{ $t('settings.userCommands') }}</h3>
    <p class="small">{{ $t('settings.userCommandsHelp') }} <i>git svn hg</i>.</p>
    <input v-model.trim="raw" class="input input--block" type="text">
  </div>
</template>

<script>
export default {
  name: 'Permissions',
  props: ['commands'],
  computed: {
    raw: {
      get() {
        return this.commands.join(' ')
      },
      set(value) {
        this.$emit('update:commands', value.split(' '))
      }
    }
  }
}
</script>
