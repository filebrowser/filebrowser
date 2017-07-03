import store from '../store/store'

function fetch (url) {
  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open('GET', `${store.state.baseURL}/api/resource${url}`, true)
    request.setRequestHeader('Authorization', `Bearer ${store.state.jwt}`)

    request.onload = () => {
      if (request.status === 200) {
        let req = JSON.parse(request.responseText)
        store.commit('updateRequest', req)
        document.title = req.name
        resolve()
      } else {
        reject()
      }
    }
    request.onerror = () => reject()
    request.send()
  })
}

export default {
  fetch
}
