<template>
  <div class="t-container">
    <span>{{ message }}</span>
    <button v-if="isReport" class="action" @click.stop="clicked">
      {{ reportText }}
    </button>
  </div>
</template>

<script>
export default {
  name: "error-toast",
  props: ["message", "reportText", "isReport"],
  methods: {
    clicked() {
      window.open(
        "https://github.com/filebrowser/filebrowser/issues/new/choose"
      );
    },
  },
};
</script>

<style>
.t-container {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.action {
  height: 40px;
  padding: 5px 10px;
  border-radius: 5px;
  color: white;
  cursor: pointer;
  border: thin solid currentColor;
}
</style>
