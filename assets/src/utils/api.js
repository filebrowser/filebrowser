import store from '@/store'

const ssl = (window.location.protocol === 'https:')

export function removePrefix (url) {
  if (url.startsWith('/files')) {
    return url.slice(6)
  }

  return url
}

export function fetch (url) {
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
          reject(new Error(request.status))
          break
      }
    }
    request.onerror = (error) => reject(error)
    request.send()
  })
}

export function rm (url) {
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

export function post (url, content = '') {
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

export function put (url, content = '') {
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

export function moveCopy (items, copy = false) {
  let promises = []

  for (let item of items) {
    let from = removePrefix(item.from)
    let to = removePrefix(item.to)

    promises.push(new Promise((resolve, reject) => {
      let request = new window.XMLHttpRequest()
      request.open('PATCH', `${store.state.baseURL}/api/resource${from}`, true)
      request.setRequestHeader('Authorization', `Bearer ${store.state.jwt}`)
      request.setRequestHeader('Destination', to)

      if (copy) {
        request.setRequestHeader('Action', 'copy')
      }

      request.onload = () => {
        if (request.status === 200) {
          resolve(request.responseText)
        } else {
          reject(request.responseText)
        }
      }

      request.onerror = (error) => reject(error)
      request.send()
    }))
  }

  return Promise.all(promises)
}

export function move (items) {
  return moveCopy(items)
}

export function copy (items) {
  return moveCopy(items, true)
}

export function checksum (url, algo) {
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

export function command (url, command, onmessage, onclose) {
  let protocol = (ssl ? 'wss:' : 'ws:')
  url = removePrefix(url)
  url = `${protocol}//${window.location.host}${store.state.baseURL}/api/command${url}`

  let conn = new window.WebSocket(url)
  conn.onopen = () => conn.send(command)
  conn.onmessage = onmessage
  conn.onclose = onclose
}

export function search (url, search, onmessage, onclose) {
  let protocol = (ssl ? 'wss:' : 'ws:')
  url = removePrefix(url)
  url = `${protocol}//${window.location.host}${store.state.baseURL}/api/search${url}`

  let conn = new window.WebSocket(url)
  conn.onopen = () => conn.send(search)
  conn.onmessage = onmessage
  conn.onclose = onclose
}

export function download (format, ...files) {
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

  if (format !== null) {
    url += `&format=${format}`
  }

  window.open(url)
}

export function getUsers () {
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

export function getUser (id) {
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

export function newUser (user) {
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

export function updateUser (user, which) {
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
    request.send(JSON.stringify({
      what: 'user',
      which: (typeof which === 'string') ? which : 'all',
      data: user
    }))
  })
}

export function deleteUser (id) {
  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open('DELETE', `${store.state.baseURL}/api/users/${id}`, true)
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
    request.send()
  })
}

export function getCommands () {
  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open('GET', `${store.state.baseURL}/api/commands/`, true)
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

export function updateCommands (commands) {
  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open('PUT', `${store.state.baseURL}/api/commands/`, true)
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
    request.send(JSON.stringify(commands))
  })
}

export function getPlugins () {
  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open('GET', `${store.state.baseURL}/api/plugins/`, true)
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

export function updatePlugins (data) {
  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open('PUT', `${store.state.baseURL}/api/plugins/`, true)
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
    request.send(JSON.stringify(data))
  })
}

export default {
  delete: rm,
  fetch,
  checksum,
  move,
  put,
  copy,
  post,
  command,
  search,
  download,
  getUser,
  newUser,
  updateUser,
  getUsers,
  getCommands,
  updateCommands,
  removePrefix,
  getPlugins,
  updatePlugins,
  deleteUser
}
