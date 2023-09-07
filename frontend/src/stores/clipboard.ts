import { defineStore } from "pinia";

export const useClipboardStore = defineStore("clipboard", {
  // convert to a function
  state: () => ({
    key: "",
    items: [],
    path: undefined,
  }),
  getters: {
    // user and jwt getter removed, no longer needed
  },
  actions: {
    // no context as first argument, use `this` instead
    updateClipboard(value: any) {
      this.key = value.key;
      this.items = value.items;
      this.path = value.path;
    },
    // easily reset state using `$reset`
    resetClipboard() {
      this.$reset();
    },
  },
});
