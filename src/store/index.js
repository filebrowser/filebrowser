import Vue from 'vue'
import Vuex from 'vuex'
import mutations from './mutations'
import getters from './getters'

Vue.use(Vuex)

const state = {
  user: null,
  req: {},
  oldReq: {},
  clipboard: {
    key: '',
    items: []
  },
  jwt: '',
  progress: 0,
  loading: false,
  reload: false,
  selected: [],
  multiple: false,
  show: null,
  showShell: false,
  showMessage: null,
  showConfirm: null
}

export default new Vuex.Store({
  strict: true,
  state,
  getters,
  mutations
})
