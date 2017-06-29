<template>
    <button @click="show" title="Info" aria-label="Info" class="action">
        <i class="material-icons">info</i>
        <span>Info</span>
    </button>
</template>

<script>
export default {
  name: 'info-button',
  methods: {
    show: function (event) {
      window.info.showInfo = true
    }
  }
}
</script>
