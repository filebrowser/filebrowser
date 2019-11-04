const getters = {
  isLogged: state => state.user !== null,
  isFiles: state => !state.loading && state.route.name === 'Files',
  isListing: (state, getters) => getters.isFiles && state.req.isDir,
  isEditor: (state, getters) => getters.isFiles && (state.req.type === 'text' || state.req.type === 'textImmutable'),
  isDiagrams: state => getters.isFiles && (state.req.type === 'bpmn'), // Only required for DiagramsViewer integration.
  selectedCount: state => state.selected.length
}

export default getters
