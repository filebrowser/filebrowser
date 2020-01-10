<template>
  <button @click="show" :aria-label="$t('buttons.edit')" :title="$t('buttons.edit')" class="action" id="edit-button">
    <i class="material-icons">mode_edit</i>
    <span>{{ $t('buttons.edit') }}</span>
  </button>
</template>

<script>
export default {
  name: 'edit-button',
  methods: {
    show: function () {
      this.$store.commit('toggleEditor')
    }
  }
}
</script>
