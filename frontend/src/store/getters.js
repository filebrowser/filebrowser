const getters = {
  isLogged: (state) => state.user !== null,
  isFiles: (state) => !state.loading && state.route.name === "Files",
  isListing: (state, getters) => getters.isFiles && state.req.isDir,
  isVisibleContext: (state, getters) =>
    getters.isListing && state.contextMenu !== null,
  selectedCount: (state) => state.selected.length,
  progress: (state) => {
    if (state.upload.progress.length == 0) {
      return 0;
    }

    let sum = state.upload.progress.reduce((acc, val) => acc + val);
    return Math.ceil((sum / state.upload.size) * 100);
  },
  onlyArchivesSelected: (state, getters) => {
    let extensions = [".zip", ".tar", ".gz", ".bz2", ".xz", ".lz4", ".sz"];
    let items = state.req.items;
    if (getters.selectedCount < 1) {
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
};

export default getters;
