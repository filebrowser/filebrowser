<template>
  <button :class="classes" @click="$emit('click')">
    <slot></slot>
  </button>
</template>

<script>
export default {
  name: "Button",

  props: {
    type: {
      type: String,
      default: "primary",
      validator: (value) => ["primary", "secondary", "link"].includes(value),
    },

    size: {
      type: String,
      default: "normal",
      validator: (value) => ["none", "normal", "large"].includes(value),
    },

    fullWidth: {
      type: Boolean,
      default: false,
    },
  },

  computed: {
    classes() {
      return [
        `app-button app-button--${this.type} app-button--${this.size}`,
        { "app-button--full-width": this.fullWidth },
        { "app-button--loading": this.loading },
      ];
    },
  },
};
</script>

<style scoped>
.app-button {
  border: 0;
  outline: 0;
  cursor: pointer;
  padding: 0.5em 1em;
  border-radius: 0.1em;
  transition: background-color 0.3s ease;
  border: 1px solid rgba(0, 0, 0, 0.05);
  box-shadow: 0 0 5px rgba(0, 0, 0, 0.05);
}

.app-button:hover {
  background-color: var(--dark-blue);
}

.app-button--primary {
  color: #ffffff;
  background-color: var(--blue);
}

.app-button--normal {
  height: 3rem;
  border-radius: 0.5rem;
}

.app-button--full-width {
  width: 100%;
}
</style>
