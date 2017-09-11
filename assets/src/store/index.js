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
  css: (() => {
    let css = window.CSS
    window.CSS = null
    return css
  })(),
  recaptcha: document.querySelector('meta[name="recaptcha"]').getAttribute('content'),
  staticGen: document.querySelector('meta[name="staticgen"]').getAttribute('content'),
  baseURL: document.querySelector('meta[name="base"]').getAttribute('content'),
  noAuth: (document.querySelector('meta[name="noauth"]').getAttribute('content') === 'true'),
  version: document.querySelector('meta[name="version"]').getAttribute('content'),
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
