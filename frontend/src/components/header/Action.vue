<template>
    <button @click="$emit('action')" :aria-label="label" :title="label" class="action">
    <i class="material-icons">{{ icon }}</i>
    <span>{{ label }}</span>
  </button>
</template>

<script>
export default {
  name: 'action',
  props: [ 'icon', 'label' ]
}
</script>

<style>

</style>