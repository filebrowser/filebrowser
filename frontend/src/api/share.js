import { fetchURL, fetchJSON, removePrefix } from './utils'

export async function list() {
  return fetchJSON('/api/shares')
}

export async function getHash(hash, shared_code = "") {
  return fetchJSON(`/api/public/share/${hash}`, {
    headers: { 'X-SHARED-CODE': shared_code },
  })
}

export async function get(url) {
  url = removePrefix(url)
  return fetchJSON(`/api/share${url}`)
}

export async function remove(hash) {
  const res = await fetchURL(`/api/share/${hash}`, {
    method: 'DELETE'
  })

  if (res.status !== 200) {
    throw new Error(res.status)
  }
}

export async function create(url, shared_code = '', expires = '', unit = 'hours') {
  url = removePrefix(url)
  url = `/api/share${url}`
  if (shared_code !== '' || expires !== '') {
    url += '?'
    var params = ''
    if (expires !== '') {
      params += `expires=${expires}&unit=${unit}`
    }
    if (shared_code !== '') {
      if (params != '') {
        params += "&"
      }
      params += `shared_code=${shared_code}`
    }
    url += params
  }
  return fetchJSON(url, {
    method: 'POST',
  })
}
