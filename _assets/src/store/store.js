import Vue from 'vue'
import Vuex from 'vuex'
import mutations from './mutations'
import getters from './getters'

Vue.use(Vuex)

const state = {
  user: {},
  req: {},
  baseURL: document.querySelector('meta[name="base"]').getAttribute('content'),
  jwt: '',
  reload: false,
  selected: [],
  multiple: false,
  prompt: null
}

export default new Vuex.Store({
  strict: process.env.NODE_ENV !== 'production',
  state,
  getters,
  mutations
})
