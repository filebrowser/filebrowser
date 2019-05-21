import { removePrefix } from './utils'
import { baseURL } from '@/utils/constants'
import store from '@/store'

const ssl = (window.location.protocol === 'https:')
const protocol = (ssl ? 'wss:' : 'ws:')

export default function command(url, command, onmessage, onclose) {
  url = removePrefix(url)
  url = `${protocol}//${window.location.host}${baseURL}/api/command${url}?auth=${store.state.jwt}`

  let conn = new window.WebSocket(url)
  conn.onopen = () => conn.send(command)
  conn.onmessage = onmessage
  conn.onclose = onclose
}
