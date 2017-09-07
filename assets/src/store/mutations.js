import i18n from '@/i18n'
import moment from 'moment'

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
    state.showConfirm = value.confirm
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
    moment.locale(value.locale)
    i18n.locale = value.locale
    state.user = value
  },
  setJWT: (state, value) => (state.jwt = value),
  multiple: (state, value) => (state.multiple = value),
  addSelected: (state, value) => (state.selected.push(value)),
  addPlugin: (state, value) => {
    state.plugins.push(value)
  },
  removeSelected: (state, value) => {
    let i = state.selected.indexOf(value)
    if (i === -1) return
    state.selected.splice(i, 1)
  },
  resetSelected: (state) => {
    state.selected = []
  },
  updateUser: (state, value) => {
    if (typeof value !== 'object') return

    for (let field in value) {
      state.user[field] = value[field]
    }
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
  },
  setSchedule: (state, value) => {
    state.schedule = value
  },
  setProgress: (state, value) => {
    state.progress = value
  }
}

export default mutations
