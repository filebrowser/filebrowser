import Vue from 'vue'
import Vuex from 'vuex'
import mutations from './mutations'
import getters from './getters'

Vue.use(Vuex)

const state = {
  user: {},
  req: {},
  clipboard: {
    key: '',
    items: []
  },
  staticGen: document.querySelector('meta[name="staticgen"]').getAttribute('content'),
  baseURL: document.querySelector('meta[name="base"]').getAttribute('content'),
  jwt: '',
  progress: 0,
  schedule: '',
  loading: false,
  reload: false,
  selected: [],
  multiple: false,
  show: null,
  showMessage: null,
  showConfirm: null
}

export default new Vuex.Store({
  strict: process.env.NODE_ENV !== 'production',
  state,
  getters,
  mutations
})
