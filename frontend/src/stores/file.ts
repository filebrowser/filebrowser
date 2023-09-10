import type { IFile } from "@/types";
import { defineStore } from "pinia";

export const useFileStore = defineStore("file", {
  // convert to a function
  state: (): {
    req: IFile | null;
    oldReq: IFile | null;
    reload: boolean;
    selected: any[];
    multiple: boolean;
    isFiles: boolean;
  } => ({
    req: null,
    oldReq: null,
    reload: false,
    selected: [],
    multiple: false,
    isFiles: false,
  }),
  getters: {
    selectedCount: (state) => state.selected.length,
    // route: () => {
    //   const routerStore = useRouterStore();
    //   return routerStore.router.currentRoute;
    // },
    // isFiles: (state) => {
    //   const layoutStore = useLayoutStore();
    //   return !layoutStore.loading && state.route._value.name === "Files";
    // },
    isListing: (state) => {
      return state.isFiles && state?.req?.isDir;
    },
  },
  actions: {
    // no context as first argument, use `this` instead
    toggleMultiple() {
      this.multiple = !this.multiple;
    },
    updateRequest(value: IFile | null) {
      this.oldReq = this.req;
      this.req = value;
    },
    removeSelected(value: any) {
      const i = this.selected.indexOf(value);
      if (i === -1) return;
      this.selected.splice(i, 1);
    },
    // easily reset state using `$reset`
    clearFile() {
      this.$reset();
    },
  },
});
