import Vue from 'vue'
import Vuex from 'vuex'
import mutations from './mutations'

Vue.use(Vuex)

const state = {
  user: window.info.user,
  req: window.info.req,
  webDavURL: window.info.webdavURL,
  baseURL: window.info.baseURL,
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
  },
  selectedCount: state => state.selected.length
}

export default new Vuex.Store({
  strict: process.env.NODE_ENV !== 'production',
  state,
  getters,
  mutations
})
