import { defineStore } from "pinia";
import { useRouterStore } from "./router";
// import { useAuthPreferencesStore } from "./auth-preferences";
// import { useAuthEmailStore } from "./auth-email";

export const useFileStore = defineStore("file", {
  // convert to a function
  state: () => ({
    req: {},
    oldReq: {},
    reload: false,
    selected: [],
    multiple: false,
  }),
  getters: {
    // user and jwt getter removed, no longer needed
    selectedCount: (state) => state.selected.length,
    route: () => {
      const routerStore = useRouterStore();
      return routerStore.router.currentRoute;
    },
    isFiles(state) {
      return !state.loading && this.route.name === "Files";
    },
    isListing(state) {
      return this.isFiles && state.req.isDir;
    },
  },
  actions: {
    // no context as first argument, use `this` instead
    toggleMultiple() {
      this.multiple = !this.multiple;
    },
    updateRequest(value) {
      this.oldReq = this.req;
      this.req = value;
    },
    removeSelected(value) {
      let i = this.selected.indexOf(value);
      if (i === -1) return;
      this.selected.splice(i, 1);
    },
    // easily reset state using `$reset`
    clearFile() {
      this.$reset();
    },
  },
});
