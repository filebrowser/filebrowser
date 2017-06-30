import store from '../store/store'

function open (url, history) {
  // Reset info
  store.commit('resetSelected')
  store.commit('multiple', false)

  let request = new window.XMLHttpRequest()
  request.open('GET', url, true)
  request.setRequestHeader('Accept', 'application/json')

  request.onload = () => {
    if (request.status === 200) {
      let req = JSON.parse(request.responseText)
      store.commit('updateRequest', req)

      if (history) {
        window.history.pushState({
          name: req.data.name,
          url: url
        }, req.data.name, url)

        document.title = req.data.name
      }
    } else {
      console.log(request.responseText)
    }
  }

  request.onerror = (error) => { console.log(error) }
  request.send()
}

function removeLastDir (url) {
  var arr = url.split('/')
  if (arr.pop() === '') {
    arr.pop()
  }
  return (arr.join('/'))
}

export default {
  reload: () => {
    open(window.location.pathname, false)
  },
  open: (url) => {
    open(url, true)
  },
  removeLastDir: removeLastDir
}
