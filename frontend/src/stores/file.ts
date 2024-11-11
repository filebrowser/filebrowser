import { defineStore } from "pinia";

export const useFileStore = defineStore("file", {
  // convert to a function
  state: (): {
    req: Resource | null;
    oldReq: Resource | null;
    reload: boolean;
    selected: number[];
    multiple: boolean;
    isFiles: boolean;
    diskUsages: Map<string, DiskUsage>;
  } => ({
    req: null,
    oldReq: null,
    reload: false,
    selected: [],
    multiple: false,
    isFiles: false,
    diskUsages: new Map<string, DiskUsage>(),
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
    onlyArchivesSelected: (state) => {
      let extensions = [".zip", ".tar", ".gz", ".bz2", ".xz", ".lz4", ".sz"];
      let items = state.req?.items;

      if (!items) {
        return false;
      }

      if (state.selected.length < 1) {
        return false;
      }

      for (const i of state.selected) {
        let item = items[i];
        if (item.isDir || !extensions.includes(item.extension)) {
          return false;
        }
      }

      return true;
    },
  },
  actions: {
    // no context as first argument, use `this` instead
    toggleMultiple() {
      this.multiple = !this.multiple;
    },
    updateRequest(value: Resource | null) {
      const selectedItems = this.selected.map((i) => this.req?.items[i]);
      this.oldReq = this.req;
      this.req = value;

      this.selected = [];

      if (!this.req?.items) return;
      this.selected = this.req.items
        .filter((item) =>
          selectedItems.some((rItem) => rItem?.url === item.url)
        )
        .map((item) => item.index);
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
    addDiskUsage(path: string, value: DiskUsage) {
      if (path[path.length - 1] == "/") {
        path = path.slice(0, -1);
      }

      this.diskUsages.set(path, value);
    },
  },
});
