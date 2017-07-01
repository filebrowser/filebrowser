// TODO: use store
var $ = window.info

function convertURL (url) {
  return window.location.origin + url.replace($.baseURL + '/', $.webdavURL + '/')
}

function move (oldLink, newLink) {
  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()

    oldLink = convertURL(oldLink)
    newLink = newLink.replace($.baseURL + '/', $.webdavURL + '/')
    newLink = window.location.origin + newLink.substring($.baseURL.length)

    request.open('MOVE', oldLink, true)
    request.setRequestHeader('Destination', newLink)
    request.onload = () => {
      if (request.status === 201 || request.status === 204) {
        resolve()
      } else {
        reject(request.statusText)
      }
    }
    request.onerror = () => reject(request.statusText)
    request.send()
  })
}

function put (link, body, headers = {}) {
  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open('PUT', convertURL(link), true)

    for (let key in headers) {
      request.setRequestHeader(key, headers[key])
    }

    request.onload = () => {
      if (request.status === 201) {
        resolve()
      } else {
        reject(request.statusText)
      }
    }
    request.onerror = () => reject(request.statusText)
    request.send(body)
  })
}

function trash (link) {
  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open('DELETE', convertURL(link), true)
    request.onload = () => {
      if (request.status === 204) {
        resolve()
      } else {
        reject(request.statusText)
      }
    }
    request.onerror = () => reject(request.statusText)
    request.send()
  })
}

function create (link) {
  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open((link.endsWith('/') ? 'MKCOL' : 'PUT'), convertURL(link), true)
    request.onload = () => {
      if (request.status === 201) {
        resolve()
      } else {
        reject(request.statusText)
      }
    }
    request.onerror = () => reject(request.statusText)
    request.send()
  })
}

export default {
  create: create,
  trash: trash,
  put: put,
  move: move
}
