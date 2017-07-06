import Vue from 'vue'
import App from './App'
import store from './store'
import router from './router'

Vue.config.productionTip = true

/* eslint-disable no-new */
new Vue({
  el: '#app',
  store,
  router,
  template: '<App/>',
  components: { App }
})
