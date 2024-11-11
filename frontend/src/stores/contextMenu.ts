import { defineStore } from "pinia";

export const useContextMenuStore = defineStore("context-menu", {
  state: (): {
    position: ContextMenuPosition | null;
  } => ({
    position: null,
  }),
  actions: {
    show(x: number, y: number) {
      this.position = { x, y };
    },
    hide() {
      this.position = null;
    },
  },
});
