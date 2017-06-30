// The Vue build version to load with the `import` command
// (runtime-only or standalone) has been set in webpack.base.conf with an alias.
import Vue from 'vue'
import App from './App'
import store from './store/store'
// simport page from './page.js'

Vue.config.productionTip = false

var $ = (window.info || window.alert('Something is wrong, please refresh!'))

// TODO: keep this here? Maybe on app.vue?
document.title = $.req.data.name

// TODO: keep this here? Maybe on app.vue?
window.addEventListener('popstate', (event) => {
  event.preventDefault()
  event.stopPropagation()

  $.req.kind = ''
  $.selected = []
  $.multiple = false
  // TODO: find a better way to do this. Maybe on app.vue?
  window.info.showHelp = false
  window.info.showInfo = false
  window.info.showDelete = false
  window.info.showRename = false
  window.info.showMove = false

  let request = new window.XMLHttpRequest()
  request.open('GET', event.state.url, true)
  request.setRequestHeader('Accept', 'application/json')

  request.onload = () => {
    if (request.status === 200) {
      $.req = JSON.parse(request.responseText)
      document.title = event.state.name
    } else {
      console.log(request.responseText)
    }
  }

  request.onerror = (error) => { console.log(error) }
  request.send()
})

/* eslint-disable no-new */
new Vue({
  el: '#app',
  store,
  template: '<App/>',
  components: { App }
})
