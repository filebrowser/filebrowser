<template>
  <button @click="show"
    :aria-label="$t('buttons.schedule')"
    :title="$t('buttons.schedule')"
    id="schedule-button"
    class="action">
    <i class="material-icons">alarm</i>
    <span>{{ $t('buttons.schedule') }}</span>
  </button>
</template>

<script>
export default {
  name: 'schedule-button',
  methods: {
    show: function (event) {
      this.$store.commit('showHover', 'schedule')
    }
  }
}
</script>
