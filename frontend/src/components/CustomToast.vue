<template>
  <div class="t-container">
    <span class="toast-message-text">{{ message }}</span>
    <button v-if="isReport" class="action" @click.stop="clicked">
      {{ reportText }}
    </button>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  message: string;
  reportText?: string;
  isReport?: boolean;
}>();

const clicked = () => {
  window.open("https://github.com/filebrowser/filebrowser/issues/new/choose");
};
</script>

<style scoped>
.t-container {
  width: 100%;
  padding: 5px 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.toast-message-text {
  white-space: normal;
  line-height: 1.25;
  font-size: 1rem;
  text-align: left;
  margin: 0;
  display: inline;
  width: auto;
  color: inherit;
  font-weight: inherit;
}


.action {
  text-align: center;
  height: 40px;
  padding: 0 10px;
  margin-left: 20px;
  border-radius: 5px;
  color: white;
  cursor: pointer;
  border: thin solid currentColor;
}

html[dir="rtl"] .action {
  margin-left: initial;
  margin-right: 20px;
}
</style>
