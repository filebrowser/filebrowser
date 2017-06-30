import Vue from 'vue'
import Vuex from 'vuex'
import mutations from './mutations'
import getters from './getters'

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

export default new Vuex.Store({
  strict: process.env.NODE_ENV !== 'production',
  state,
  getters,
  mutations
})
