import Vue from 'vue'
import App from './App'
import store from './store'
import router from './router'
import i18n from './i18n'

Vue.config.productionTip = true

/* eslint-disable no-new */
new Vue({
  el: '#app',
  store,
  router,
  i18n,
  template: '<App/>',
  components: { App }
})
