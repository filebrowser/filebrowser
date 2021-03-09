import { fetchJSON, removePrefix } from './utils'
import { baseURL } from '@/utils/constants'

export async function fetch(hash, password = "") {
  return fetchJSON(`/api/public/share/${hash}`, {
    headers: {'X-SHARE-PASSWORD': password},
  })
}

export function download(format, hash, token, ...files) {
  let url = `${baseURL}/api/public/dl/${hash}`

  const prefix = `/share/${hash}`
  if (files.length === 1) {
    url += removePrefix(files[0], prefix) + '?'
  } else {
    let arg = ''

    for (let file of files) {
      arg += removePrefix(file, prefix) + ','
    }

    arg = arg.substring(0, arg.length - 1)
    arg = encodeURIComponent(arg)
    url += `/?files=${arg}&`
  }

  if (format) {
    url += `algo=${format}&`
  }

  if (token) {
    url += `token=${token}&`
  }

  window.open(url)
}