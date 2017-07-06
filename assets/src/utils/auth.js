import cookie from './cookie'
import store from '@/store'
import router from '@/router'

function parseToken (token) {
  document.cookie = `auth=${token}; max-age=86400; path=${store.state.baseURL}`
  let res = token.split('.')
  let user = JSON.parse(window.atob(res[1]))
  store.commit('setJWT', token)
  store.commit('setUser', user)
}

function loggedIn () {
  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open('GET', `${store.state.baseURL}/api/auth/renew`, true)
    request.setRequestHeader('Authorization', `Bearer ${cookie('auth')}`)

    request.onload = () => {
      if (request.status === 200) {
        parseToken(request.responseText)
        resolve()
      } else {
        reject()
      }
    }
    request.onerror = () => reject()
    request.send()
  })
}

function login (user, password) {
  let data = {username: user, password: password}
  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open('POST', `${store.state.baseURL}/api/auth/get`, true)

    request.onload = () => {
      if (request.status === 200) {
        parseToken(request.responseText)
        resolve()
      } else {
        reject(request.responseText)
      }
    }
    request.onerror = () => reject()
    request.send(JSON.stringify(data))
  })
}

function logout () {
  document.cookie = `auth='nothing'; max-age=0; path=${store.state.baseURL}`
  router.push({path: '/login'})
}

export default {
  loggedIn: loggedIn,
  login: login,
  logout: logout
}
