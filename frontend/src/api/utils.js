import store from '@/store'
import { renew } from '@/utils/auth'
import { baseURL } from '@/utils/constants'

export async function fetchURL (url, opts) {
  opts = opts || {}
  opts.headers = opts.headers || {}

  let { headers, ...rest } = opts

  const res = await fetch(`${baseURL}${url}`, {
    headers: {
      'X-Auth': store.state.jwt,
      ...headers
    },
    ...rest
  })

  if (res.headers.get('X-Renew-Token') === 'true') {
    await renew(store.state.jwt)
  }

  return res
}

export async function fetchJSON (url, opts) {
  const res = await fetchURL(url, opts)

  if (res.status === 200) {
    return res.json()
  } else {
    throw new Error(res.status)
  }
}

export function removePrefix (url) {
  if (url.startsWith('/files')) {
    url = url.slice(6)
  }

  if (url === '') url = '/'
  if (url[0] !== '/') url = '/' + url
  return url
}

