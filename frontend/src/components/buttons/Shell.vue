<template>
  <button @click="show" :aria-label="$t('buttons.shell')" :title="$t('buttons.shell')" class="action">
    <i class="material-icons">code</i>
    <span>{{ $t('buttons.shell') }}</span>
  </button>
</template>

<script>
export default {
  name: 'shell-button',
  methods: {
    show: function () {
      this.$store.commit('toggleShell')
    }
  }
}
</script>
