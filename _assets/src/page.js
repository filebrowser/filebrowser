function open (url, history) {
  window.info.page.kind = ''

  let request = new window.XMLHttpRequest()
  request.open('GET', url, true)
  request.setRequestHeader('Accept', 'application/json')

  request.onload = () => {
    if (request.status === 200) {
      window.info.page = JSON.parse(request.responseText)

      if (history) {
        window.history.pushState({
          name: window.info.page.name,
          url: url
        }, window.info.page.name, url)

        document.title = window.info.page.name
      }
    } else {
      console.log(request.responseText)
    }
  }

  request.onerror = (error) => { console.log(error) }
  request.send()
}

export default {
  reload: () => {
    open(window.location.pathname, false)
  },
  open: (url) => {
    open(url, true)
  }
}
