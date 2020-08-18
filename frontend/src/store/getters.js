const getters = {
  isLogged: state => state.user !== null,
  isFiles: state => !state.loading && state.route.name === 'Files',
  isListing: (state, getters) => getters.isFiles && state.req.isDir,
  isEditor: (state, getters) => getters.isFiles && (state.req.type === 'text' || state.req.type === 'textImmutable'),
  isPreview: state => state.previewMode,
  selectedCount: state => state.selected.length,
  progress : state => {
    if (state.upload.progress.length == 0) {
      return 0;
    }

    let sum = state.upload.progress.reduce((acc, val) => acc + val)
    return Math.ceil(sum / state.upload.size * 100);
  }
}

export default getters
