const getters = {
  isLogged: state => state.user !== null,
  isFiles: state => !state.loading && state.route.name === 'Files',
  isListing: (state, getters) => getters.isFiles && state.req.isDir,
  isPreview: (state, getters) => !state.loading && !getters.isListing && !getters.isEditor,
  isEditor: (state, getters) => getters.isFiles && state.showEditor,
  isFileEditable: (state) => state.req.type === 'text' || state.req.type === 'textImmutable',
  getPreviewContent: state => state.previewContent,
  selectedCount: state => state.selected.length
}

export default getters
