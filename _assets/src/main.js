// The Vue build version to load with the `import` command
// (runtime-only or standalone) has been set in webpack.base.conf with an alias.
import Vue from 'vue'
import App from './App'
// simport page from './page.js'

Vue.config.productionTip = false

var $ = (window.info || window.alert('Something is wrong, please refresh!'))

// TODO: keep this here?
document.title = $.req.name

// TODO: keep this here?
window.addEventListener('popstate', (event) => {
  event.preventDefault()
  event.stopPropagation()

  $.req.kind = ''
  $.listing.selected.length = 0
  $.listing.selected.multiple = false

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
  template: '<App/>',
  components: { App }
})
