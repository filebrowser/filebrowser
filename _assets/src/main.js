// The Vue build version to load with the `import` command
// (runtime-only or standalone) has been set in webpack.base.conf with an alias.
import Vue from 'vue'
import App from './App'
// simport page from './page.js'

Vue.config.productionTip = false

window.info = (window.info || window.alert('Something is wrong, please refresh!'))
window.ssl = (window.location.protocol === 'https:')

// TODO: keep this here?
document.title = window.info.page.name

// TODO: keep this here?
window.addEventListener('popstate', (event) => {
  event.preventDefault()
  event.stopPropagation()

  window.info.page.kind = ''

  let request = new window.XMLHttpRequest()
  request.open('GET', event.state.url, true)
  request.setRequestHeader('Accept', 'application/json')

  request.onload = () => {
    if (request.status === 200) {
      window.info.page = JSON.parse(request.responseText)
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
