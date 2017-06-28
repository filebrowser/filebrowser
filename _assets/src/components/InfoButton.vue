<template>
    <button title="Info" aria-label="Info" class="action" id="info">
        <i class="material-icons">info</i>
        <span>Info</span>
    </button>
</template>

<script>
export default {
  name: 'info-button',
  data: function () {
    return window.info.page.data
  }
}
</script>
