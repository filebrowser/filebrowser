<template>
  <input :type="type" class="app-input" @input="handleInput" />
</template>

<script>
export default {
  name: "Input",

  props: {
    type: {
      type: String,
      default: "text",
    },
  },

  methods: {
    handleInput(e) {
      this.$emit("input", e.target.value);
    },
  },
};
</script>

<style scoped>
.app-input {
  margin: 0;
  height: 48px;
  outline: none;
  padding: 0 1rem;
  color: #333333;
  border-radius: 0.5rem;
  transition: border 0.3s ease;
  border: 1px solid rgba(0, 0, 0, 0.1);
}

.app-input:is(:hover, :focus) {
  border: 1px solid rgba(0, 0, 0, 0.3);
}
</style>
