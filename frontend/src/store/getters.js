const getters = {
  isLogged: (state) => state.user !== null,
  isFiles: (state) => !state.loading && state.route.name === "Files",
  isListing: (state, getters) => getters.isFiles && state.req.isDir,
  selectedCount: (state) => state.selected.length,
  progress: (state) => {
    if (state.upload.progress.length == 0) {
      return 0;
    }

    let totalSize = state.upload.sizes.reduce((a, b) => a + b, 0);

    let sum = state.upload.progress.reduce((acc, val) => acc + val);
    return Math.ceil((sum / totalSize) * 100);
  },
  filesInUploadCount: (state) => {
    let total =
      Object.keys(state.upload.uploads).length + state.upload.queue.length;
    return total;
  },
  filesInUpload: (state) => {
    let files = [];

    for (let index in state.upload.uploads) {
      let upload = state.upload.uploads[index];
      let id = upload.id;
      let type = upload.type;
      let name = decodeURIComponent(upload.path.replace(/^.*[\\/]/, ""));
      let progress = state.upload.progress[id];
      let size = state.upload.sizes[id];

      files.push({
        id,
        name,
        progress: Math.ceil((progress / size) * 100),
        type,
      });
    }

    return files.sort((a, b) => a.progress - b.progress);
  },
};

export default getters;
