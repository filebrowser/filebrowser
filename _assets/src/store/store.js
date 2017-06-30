import Vue from 'vue'
import Vuex from 'vuex'

Vue.use(Vuex)

const state = {
  ssl: (window.location.protocol === 'https:'),
  selected: [],
  multiple: false,
  showInfo: false,
  showHelp: false,
  showDelete: false,
  showRename: false,
  showMove: false,
  showNewFile: false,
  showNewDir: false
}

const getters = {
  showOverlay: state => {
    return state.showInfo ||
      state.showHelp ||
      state.showDelete ||
      state.showRename ||
      state.showMove ||
      state.showNewFile ||
      state.showNewDir
  }
}

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
  resetSelected: (state) => {
    state.selected.length = 0
  }
}

export default new Vuex.Store({
  strict: process.env.NODE_ENV !== 'production',
  state,
  getters,
  mutations
})
