import { fetchURL, fetchJSON } from './utils'

export function get () {
  return fetchJSON(`/api/settings`, {})
}

export async function update (settings) {
  const res = await fetchURL(`/api/settings`, {
    method: 'PUT',
    body: JSON.stringify(settings)
  })

  if (res.status !== 200) {
    throw new Error(res.status)
  }
}
