import { fetchURL, removePrefix } from './utils'
import { baseURL } from '@/utils/constants'
import store from '@/store'

export async function fetch (url) {
  url = removePrefix(url)

  const res = await fetchURL(`/api/resources${url}`, {})

  if (res.status === 200) {
    let data = await res.json()
    data.url = `/files${data.path}`

    if (data.isDir) {
      if (!data.url.endsWith('/')) data.url += '/'
      data.items = data.items.map((item, index) => {
        item.index = index
        item.url = `${data.url}${encodeURIComponent(item.name)}`

        if (item.isDir) {
          item.url += '/'
        }

        return item
      })
    }

    return data
  } else {
    throw new Error(res.status)
  }
}

async function resourceAction (url, method, content) {
  url = removePrefix(url)

  let opts = { method }

  if (content) {
    opts.body = content
  }

  const res = await fetchURL(`/api/resources${url}`, opts)

  if (res.status !== 200) {
    throw new Error(res.responseText)
  } else {
    return res
  }
}

export async function remove (url) {
  return resourceAction(url, 'DELETE')
}

export async function put (url, content = '') {
  return resourceAction(url, 'PUT', content)
}

export function download (format, ...files) {
  let url = `${baseURL}/api/raw`

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
    url += `algo=${format}&`
  }

  url += `auth=${store.state.jwt}`
  window.open(url)
}

export async function post (url, content = '', overwrite = false, onupload) {
  url = removePrefix(url)

  return new Promise((resolve, reject) => {
    let request = new XMLHttpRequest()
    request.open('POST', `${baseURL}/api/resources${url}?override=${overwrite}`, true)
    request.setRequestHeader('Authorization', `Bearer ${store.state.jwt}`)

    if (typeof onupload === 'function') {
      request.upload.onprogress = onupload
    }

    request.onload = () => {
      if (request.status === 200) {
        resolve(request.responseText)
      } else if (request.status === 409) {
        reject(request.status)
      } else {
        reject(request.responseText)
      }
    }

    request.onerror = (error) => {
      reject(error)
    }

    request.send(content)
  })
}

function moveCopy (items, copy = false) {
  let promises = []

  for (let item of items) {
    let from = removePrefix(item.from)
    let to = encodeURIComponent(removePrefix(item.to))
    let url = `${from}?action=${copy ? 'copy' : 'rename'}&destination=${to}`
    promises.push(resourceAction(url, 'PATCH'))
  }

  return Promise.all(promises)
}

export function move (items) {
  return moveCopy(items)
}

export function copy (items) {
  return moveCopy(items, true)
}

export async function checksum (url, algo) {
  const data = await resourceAction(`${url}?checksum=${algo}`, 'GET')
  return (await data.json()).checksums[algo]
}
