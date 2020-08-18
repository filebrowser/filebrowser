<template>
  <button :title="$t('buttons.info')" :aria-label="$t('buttons.info')" class="action" @click="$emit('change-size')">
    <i class="material-icons">{{ this.icon }}</i>
    <span>{{ $t('buttons.info') }}</span>
  </button>
</template>

<script>
export default {
  name: 'preview-size-button',
  props: [ 'size' ],
  computed: {
    icon () {
      if (this.size) {
        return 'photo_size_select_large'
      }

      return 'hd'
    }
  }
}
</script>
