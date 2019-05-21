import { fetchJSON } from './utils'

export async function get () {
  return await fetchJSON(`/api/context`, {})
}
