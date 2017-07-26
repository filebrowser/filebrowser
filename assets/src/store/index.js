import Vue from 'vue'
import Vuex from 'vuex'
import mutations from './mutations'
import getters from './getters'

Vue.use(Vuex)

const state = {
  user: {},
  req: {},
  plugins: window.plugins || [],
  clipboard: {
    key: '',
    items: []
  },
  baseURL: document.querySelector('meta[name="base"]').getAttribute('content'),
  jwt: '',
  loading: false,
  reload: false,
  selected: [],
  multiple: false,
  show: null,
  showMessage: null
}

export default new Vuex.Store({
  strict: process.env.NODE_ENV !== 'production',
  state,
  getters,
  mutations
})
