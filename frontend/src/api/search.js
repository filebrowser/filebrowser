import { fetchJSON, removePrefix } from './utils'

export default async function search (url, query) {
  url = removePrefix(url)
  query = encodeURIComponent(query)

  return fetchJSON(`/api/search${url}?query=${query}`, {})
}
