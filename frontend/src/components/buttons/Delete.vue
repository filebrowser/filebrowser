<template>
  <button id="delete-button" :aria-label="$t('buttons.delete')" :title="$t('buttons.delete')" class="action" @click="show">
    <i class="material-icons">delete</i>
    <span>{{ $t('buttons.delete') }}</span>
  </button>
</template>

<script>
export default {
  name: 'DeleteButton',
  methods: {
    show: function() {
      this.$store.commit('showHover', 'delete')
    }
  }
}
</script>
