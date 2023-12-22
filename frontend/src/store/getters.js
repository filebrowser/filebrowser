const getters = {
  isLogged: (state) => state.user !== null,
  isFiles: (state) => !state.loading && state.route.name === "Files",
  isListing: (state, getters) => getters.isFiles && state.req.isDir,
  isVisibleContext: (state, getters) =>
    getters.isListing && state.contextMenu !== null,
  selectedCount: (state) => state.selected.length,
  progress: (state) => {
    if (state.upload.progress.length === 0) {
      return 0;
    }

    let totalSize = state.upload.sizes.reduce((a, b) => a + b, 0);

    let sum = state.upload.progress.reduce((acc, val) => acc + val);
    return Math.ceil((sum / totalSize) * 100);
  },
  filesInUploadCount: (state) => {
    return Object.keys(state.upload.uploads).length + state.upload.queue.length;
  },
  filesInUpload: (state) => {
    let files = [];

    for (let index in state.upload.uploads) {
      let upload = state.upload.uploads[index];
      let id = upload.id;
      let type = upload.type;
      let name = upload.file.name;
      let size = state.upload.sizes[id];
      let isDir = upload.file.isDir;
      let progress = isDir
        ? 100
        : Math.ceil((state.upload.progress[id] / size) * 100);

      files.push({
        id,
        name,
        progress,
        type,
        isDir,
      });
    }

    return files.sort((a, b) => a.progress - b.progress);
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
  currentPrompt: (state) => {
    return state.prompts.length > 0
      ? state.prompts[state.prompts.length - 1]
      : null;
  },
  currentPromptName: (_, getters) => {
    return getters.currentPrompt?.prompt;
  },
  uploadSpeed: (state) => state.upload.speedMbyte,
  eta: (state) => state.upload.eta,
};

export default getters;
