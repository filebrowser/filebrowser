import cookie from './cookie'
import store from '@/store'
import router from '@/router'
import { Base64 } from 'js-base64'

function parseToken (token) {
  let path = store.state.baseURL
  if (path === '') path = '/'
  document.cookie = `auth=${token}; max-age=86400; path=${path}`
  let res = token.split('.')
  let user = JSON.parse(Base64.decode(res[1]))
  if (!user.commands) {
    user.commands = []
  }

  store.commit('setJWT', token)
  store.commit('setUser', user)
}

function loggedIn () {
  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open('GET', `${store.state.baseURL}/api/auth/renew`, true)
    if (!store.state.noAuth) request.setRequestHeader('Authorization', `Bearer ${cookie('auth')}`)

    request.onload = () => {
      if (request.status === 200) {
        parseToken(request.responseText)
        resolve()
      } else {
        reject(new Error(request.responseText))
      }
    }
    request.onerror = () => reject(new Error('Could not finish the request'))
    request.send()
  })
}

function login (user, password, captcha) {
  let data = {username: user, password: password, recaptcha: captcha}
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
    request.onerror = () => reject(new Error('Could not finish the request'))
    request.send(JSON.stringify(data))
  })
}

function logout () {
  let path = store.state.baseURL
  if (path === '') path = '/'
  document.cookie = `auth='nothing'; max-age=0; path=${path}`
  router.push({path: '/login'})
}

export default {
  loggedIn: loggedIn,
  login: login,
  logout: logout
}
