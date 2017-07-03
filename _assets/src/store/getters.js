const getters = {
  showInfo: state => (state.prompt === 'info'),
  showHelp: state => (state.prompt === 'help'),
  showDelete: state => (state.prompt === 'delete'),
  showRename: state => (state.prompt === 'rename'),
  showMove: state => (state.prompt === 'move'),
  showNewFile: state => (state.prompt === 'newFile'),
  showNewDir: state => (state.prompt === 'newDir'),
  showDownload: state => (state.prompt === 'download'),
  showOverlay: state => (state.prompt !== null),
  selectedCount: state => state.selected.length
}

export default getters
