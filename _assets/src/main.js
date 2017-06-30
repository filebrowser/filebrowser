import Vue from 'vue'
import App from './App'
import store from './store/store'

Vue.config.productionTip = false

if (window.info === undefined || window.info === null) {
  window.alert('Something is wrong, please refresh!')
  window.location.reload()
}

/* eslint-disable no-new */
new Vue({
  el: '#app',
  store,
  template: '<App/>',
  components: { App }
})
