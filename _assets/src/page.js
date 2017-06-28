var $ = window.info

function open (url, history) {
  // Reset info
  $.listing.selected.length = 0
  $.listing.selected.multiple = false
  $.req.kind = ''

  let request = new window.XMLHttpRequest()
  request.open('GET', url, true)
  request.setRequestHeader('Accept', 'application/json')

  request.onload = () => {
    if (request.status === 200) {
      $.req = JSON.parse(request.responseText)

      if (history) {
        window.history.pushState({
          name: $.req.name,
          url: url
        }, $.req.name, url)

        document.title = $.req.name
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
