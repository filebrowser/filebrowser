import store from '@/store'

const ssl = (window.location.protocol === 'https:')

function removePrefix (url) {
  if (url.startsWith('/files')) {
    return url.slice(6)
  }

  return url
}

function fetch (url) {
  url = removePrefix(url)

  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open('GET', `${store.state.baseURL}/api/resource${url}`, true)
    request.setRequestHeader('Authorization', `Bearer ${store.state.jwt}`)

    request.onload = () => {
      switch (request.status) {
        case 200:
          resolve(JSON.parse(request.responseText))
          break
        default:
          reject({
            message: request.responseText,
            status: request.status
          })
          break
      }
    }
    request.onerror = (error) => reject(error)
    request.send()
  })
}

function rm (url) {
  url = removePrefix(url)

  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open('DELETE', `${store.state.baseURL}/api/resource${url}`, true)
    request.setRequestHeader('Authorization', `Bearer ${store.state.jwt}`)

    request.onload = () => {
      if (request.status === 200) {
        resolve(request.responseText)
      } else {
        reject(request.responseText)
      }
    }

    request.onerror = (error) => reject(error)
    request.send()
  })
}

function post (url, content = '') {
  url = removePrefix(url)

  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open('POST', `${store.state.baseURL}/api/resource${url}`, true)
    request.setRequestHeader('Authorization', `Bearer ${store.state.jwt}`)

    request.onload = () => {
      if (request.status === 200) {
        resolve(request.responseText)
      } else {
        reject(request.responseText)
      }
    }

    request.onerror = (error) => reject(error)
    request.send(content)
  })
}

function put (url, content = '') {
  url = removePrefix(url)

  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open('PUT', `${store.state.baseURL}/api/resource${url}`, true)
    request.setRequestHeader('Authorization', `Bearer ${store.state.jwt}`)

    request.onload = () => {
      if (request.status === 200) {
        resolve(request.responseText)
      } else {
        reject(request.responseText)
      }
    }

    request.onerror = (error) => reject(error)
    request.send(content)
  })
}

function move (oldLink, newLink) {
  oldLink = removePrefix(oldLink)
  newLink = removePrefix(newLink)

  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open('PATCH', `${store.state.baseURL}/api/resource${oldLink}`, true)
    request.setRequestHeader('Authorization', `Bearer ${store.state.jwt}`)
    request.setRequestHeader('Destination', newLink)

    request.onload = () => {
      if (request.status === 200) {
        resolve(request.responseText)
      } else {
        reject(request.responseText)
      }
    }

    request.onerror = (error) => reject(error)
    request.send()
  })
}

function checksum (url, algo) {
  url = removePrefix(url)

  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open('GET', `${store.state.baseURL}/api/checksum${url}?algo=${algo}`, true)
    request.setRequestHeader('Authorization', `Bearer ${store.state.jwt}`)

    request.onload = () => {
      if (request.status === 200) {
        resolve(request.responseText)
      } else {
        reject(request.responseText)
      }
    }
    request.onerror = (error) => reject(error)
    request.send()
  })
}

function command (url, command, onmessage, onclose) {
  let protocol = (ssl ? 'wss:' : 'ws:')
  url = removePrefix(url)
  url = `${protocol}//${window.location.hostname}${store.state.baseURL}/api/command${url}?token=${store.state.jwt}`

  let conn = new window.WebSocket(url)
  conn.onopen = () => conn.send(command)
  conn.onmessage = onmessage
  conn.onclose = onclose
}

function search (url, search, onmessage, onclose) {
  let protocol = (ssl ? 'wss:' : 'ws:')
  url = removePrefix(url)
  url = `${protocol}//${window.location.hostname}${store.state.baseURL}/api/search${url}?token=${store.state.jwt}`

  let conn = new window.WebSocket(url)
  conn.onopen = () => conn.send(search)
  conn.onmessage = onmessage
  conn.onclose = onclose
}

function download (format, ...files) {
  let url = `${store.state.baseURL}/api/download`

  if (files.length === 1) {
    url += removePrefix(files[0]) + '?'
  } else {
    let arg = ''

    for (let file of files) {
      arg += removePrefix(file) + ','
    }

    arg = arg.substring(0, arg.length - 1)
    arg = encodeURIComponent(arg)
    url += `/?files=${arg}&`
  }

  url += `token=${store.state.jwt}`

  if (format !== null) {
    url += `&format=${format}`
  }

  window.open(url)
}

function getUsers () {
  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open('GET', `${store.state.baseURL}/api/users/`, true)
    request.setRequestHeader('Authorization', `Bearer ${store.state.jwt}`)

    request.onload = () => {
      switch (request.status) {
        case 200:
          resolve(JSON.parse(request.responseText))
          break
        default:
          reject(request.responseText)
          break
      }
    }
    request.onerror = (error) => reject(error)
    request.send()
  })
}

function getUser (id) {
  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open('GET', `${store.state.baseURL}/api/users/${id}`, true)
    request.setRequestHeader('Authorization', `Bearer ${store.state.jwt}`)

    request.onload = () => {
      switch (request.status) {
        case 200:
          resolve(JSON.parse(request.responseText))
          break
        default:
          reject(request.responseText)
          break
      }
    }
    request.onerror = (error) => reject(error)
    request.send()
  })
}

function newUser (user) {
  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open('POST', `${store.state.baseURL}/api/users/`, true)
    request.setRequestHeader('Authorization', `Bearer ${store.state.jwt}`)

    request.onload = () => {
      switch (request.status) {
        case 201:
          resolve(request.getResponseHeader('Location'))
          break
        default:
          reject(request.responseText)
          break
      }
    }
    request.onerror = (error) => reject(error)
    request.send(JSON.stringify(user))
  })
}

function updateUser (user) {
  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open('PUT', `${store.state.baseURL}/api/users/${user.ID}`, true)
    request.setRequestHeader('Authorization', `Bearer ${store.state.jwt}`)

    request.onload = () => {
      switch (request.status) {
        case 200:
          resolve(request.getResponseHeader('Location'))
          break
        default:
          reject(request.responseText)
          break
      }
    }
    request.onerror = (error) => reject(error)
    request.send(JSON.stringify(user))
  })
}

function updatePassword (password) {
  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open('PUT', `${store.state.baseURL}/api/users/self`, true)
    request.setRequestHeader('Authorization', `Bearer ${store.state.jwt}`)

    request.onload = () => {
      switch (request.status) {
        case 200:
          resolve()
          break
        default:
          reject(request.responseText)
          break
      }
    }
    request.onerror = (error) => reject(error)
    request.send(JSON.stringify({ 'password': password }))
  })
}

function updateCSS (css) {
  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open('PUT', `${store.state.baseURL}/api/users/self`, true)
    request.setRequestHeader('Authorization', `Bearer ${store.state.jwt}`)

    request.onload = () => {
      switch (request.status) {
        case 200:
          resolve()
          break
        default:
          reject(request.responseText)
          break
      }
    }
    request.onerror = (error) => reject(error)
    request.send(JSON.stringify({ 'css': css }))
  })
}

export default {
  delete: rm,
  fetch,
  checksum,
  move,
  put,
  post,
  command,
  search,
  download,
  getUser,
  newUser,
  updateUser,
  getUsers,
  updatePassword,
  updateCSS
}
