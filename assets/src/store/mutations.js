import i18n from '@/i18n'

const mutations = {
  closeHovers: state => {
    state.show = null
    state.showMessage = null
  },
  showHover: (state, value) => {
    if (typeof value !== 'object') {
      state.show = value
      return
    }

    state.show = value.prompt
    state.showMessage = value.message
  },
  showError: (state, value) => {
    state.show = 'error'
    state.showMessage = value
  },
  showSuccess: (state, value) => {
    state.show = 'success'
    state.showMessage = value
  },
  setLoading: (state, value) => { state.loading = value },
  setReload: (state, value) => { state.reload = value },
  setUser: (state, value) => {
    i18n.locale = value.locale
    state.user = value
  },
  setUserCSS: (state, value) => (state.user.css = value),
  setJWT: (state, value) => (state.jwt = value),
  multiple: (state, value) => (state.multiple = value),
  addSelected: (state, value) => (state.selected.push(value)),
  removeSelected: (state, value) => {
    let i = state.selected.indexOf(value)
    if (i === -1) return
    state.selected.splice(i, 1)
  },
  resetSelected: (state) => {
    state.selected = []
  },
  listingDisplay: (state, value) => {
    state.req.display = value
  },
  updateRequest: (state, value) => {
    state.req = value
  },
  updateClipboard: (state, value) => {
    state.clipboard.key = value.key
    state.clipboard.items = value.items
  },
  resetClipboard: (state) => {
    state.clipboard.key = ''
    state.clipboard.items = []
  }
}

export default mutations
