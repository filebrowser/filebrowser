<template>
  <button :aria-label="$t('buttons.shell')" :title="$t('buttons.shell')" class="action" @click="show">
    <i class="material-icons">code</i>
    <span>{{ $t('buttons.shell') }}</span>
  </button>
</template>

<script>
export default {
  name: 'ShellButton',
  methods: {
    show: function() {
      this.$store.commit('toggleShell')
    }
  }
}
</script>
