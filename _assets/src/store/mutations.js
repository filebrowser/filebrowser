const mutations = {
  showInfo: (state, value) => (state.showInfo = value),
  showHelp: (state, value) => (state.showHelp = value),
  showDelete: (state, value) => (state.showDelete = value),
  showRename: (state, value) => (state.showRename = value),
  showMove: (state, value) => (state.showMove = value),
  showNewFile: (state, value) => (state.showNewFile = value),
  showNewDir: (state, value) => (state.showNewDir = value),
  resetPrompts: (state) => {
    state.showHelp = false
    state.showInfo = false
    state.showDelete = false
    state.showRename = false
    state.showMove = false
    state.showNewFile = false
    state.showNewDir = false
  },
  multiple: (state, value) => (state.multiple = value),
  addSelected: (state, value) => (state.selected.push(value)),
  removeSelected: (state, value) => {
    let i = state.selected.indexOf(value)
    if (i === -1) return
    state.selected.splice(i, 1)
  },
  resetSelected: (state) => {
    state.selected.length = 0
  },
  updateRequest: (state, value) => {
    state.req.kind = value.kind
    state.req.data = value.data
  }
}

export default mutations
