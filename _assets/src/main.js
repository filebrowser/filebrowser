import Vue from 'vue'
import App from './App'
import store from './store/store'

Vue.config.productionTip = false

var $ = (window.info || window.alert('Something is wrong, please refresh!'))

// TODO: keep this here? Maybe on app.vue?
document.title = $.req.data.name

/* eslint-disable no-new */
new Vue({
  el: '#app',
  store,
  template: '<App/>',
  components: { App }
})
