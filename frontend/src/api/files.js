import { fetchURL, removePrefix } from './utils'
import { md5Generate } from '../../src/utils/md5'
import { baseURL } from '@/utils/constants'
import store from '@/store'

/* eslint-disable no-debugger */
export async function fetch(url) {
  url = removePrefix(url)

  const res = await fetchURL(`/api/resources${url}`, {})

  if (res.status === 200) {
    let data = await res.json()
    data.url = `/files${url}`

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

async function resourceAction(url, method, content) {
  url = removePrefix(url)

  let opts = { method }

  if (content) {
    opts.body = content
  }

  const res = await fetchURL(`/api/resources${url}`, opts)

  if (res.status !== 200) {
    throw new Error(await res.text())
  } else {
    return res
  }
}

export async function remove(url) {
  return resourceAction(url, 'DELETE')
}

export async function put(url, content = '') {
  return resourceAction(url, 'PUT', content)
}

export function download(format, ...files) {
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

export async function post(url, content = '', overwrite = false, onupload) {
  url = removePrefix(url)

  let bufferContent
  if (content instanceof Blob && !['http:', 'https:'].includes(window.location.protocol)) {
    bufferContent = await new Response(content).arrayBuffer()
  }

  let partialUpload = function partialUpload(url, params, content) {
    return new Promise((resolve, reject) => {
      debugger;
      let request = new XMLHttpRequest()
      request.open('POST', `${baseURL}/api/resources${url}?override=${overwrite}${params}`, true)
      request.setRequestHeader('X-Auth', store.state.jwt)
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

      request.send(content);
    })
  }

  if (!content) {
    //create folder or create new file
    await partialUpload(url, "", content)
    return;
  }

  debugger;
  let fileSize = content.size;
  let mb = 1024 * 1024;
  let fileSizeMB = fileSize / mb;
  let chunkSize = fileSizeMB <= 50 ? 5 * mb : 10 * mb;//each chunck capacity
  let totalChunks = Math.ceil(fileSize / chunkSize);//get total chunck pieces
  let fileID = await md5Generate(content);
  let allContent = bufferContent || content;
  let tryCount = 0;
  let blobSlice = File.prototype.slice || File.prototype.mozSlice || File.prototype.webkitSlice
  for (let index = 0; index < totalChunks; index++) {
    let fileContent = null;
    let startIndex, endInex;
    if (index < totalChunks - 1) {
      startIndex = index * chunkSize;
      endInex = (index + 1) * chunkSize;
    } else {
      startIndex = index * chunkSize;
      endInex = fileSize;
    }
    fileContent = blobSlice.call(allContent, startIndex, endInex);
    
    debugger;
    let params = `&fileID=${fileID}&chunckIndex=${index + 1}&totalChunck=${totalChunks}`
    await partialUpload(url, params, fileContent).catch(err => {
      debugger;
      if (tryCount <= 2) {//one file try three times
        index--;
        tryCount++;
      } else {
        throw err;
      }
    });
    tryCount = 0;
  }
}

function moveCopy(items, copy = false, overwrite = false, rename = false) {
  let promises = []

  for (let item of items) {
    const from = item.from
    const to = encodeURIComponent(removePrefix(item.to))
    const url = `${from}?action=${copy ? 'copy' : 'rename'}&destination=${to}&override=${overwrite}&rename=${rename}`
    promises.push(resourceAction(url, 'PATCH'))
  }

  return Promise.all(promises)
}

export function move(items, overwrite = false, rename = false) {
  return moveCopy(items, false, overwrite, rename)
}

export function copy(items, overwrite = false, rename = false) {
  return moveCopy(items, true, overwrite, rename)
}

export async function checksum(url, algo) {
  const data = await resourceAction(`${url}?checksum=${algo}`, 'GET')
  return (await data.json()).checksums[algo]
}
