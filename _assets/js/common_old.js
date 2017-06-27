'use strict'

var data = (window.data || window.alert('Something is wrong, please refresh!'))

/*
var tempID = '_fm_internal_temporary_id'

var templates = {}
var selectedItems = []
var overlay
var clickOverlay

// Removes an element, if exists, from an array
Array.prototype.removeElement = function (element) {
  var i = this.indexOf(element)
  if (i !== -1) {
    this.splice(i, 1)
  }
}

// Replaces an element inside an array by another
Array.prototype.replaceElement = function (oldElement, newElement) {
  var i = this.indexOf(oldElement)
  if (i !== -1) {
    this[i] = newElement
  }
}

// Sends a costum event to itself
Document.prototype.sendCostumEvent = function (text) {
  this.dispatchEvent(new window.CustomEvent(text))
}

// Gets the content of a cookie
Document.prototype.getCookie = function (name) {
  var re = new RegExp('(?:(?:^|.*;\\s*)' + name + '\\s*\\=\\s*([^;]*).*$)|^.*$')
  return document.cookie.replace(re, '$1')
}


function getCSSRule (rules) {
  for (let i = 0; i < rules.length; i++) {
    rules[i] = rules[i].toLowerCase()
  }

  let result = null
  let find = Array.prototype.find

  find.call(document.styleSheets, styleSheet => {
    result = find.call(styleSheet.cssRules, cssRule => {
      let found = false

      if (cssRule instanceof CSSStyleRule) {
        for (let i = 0; i < rules.length; i++) {
          if (cssRule.selectorText.toLowerCase() === rules[i]) {
            found = true
          }
        }
      }

      return found
    })

    return result != null
  })

  return result
}

/* * * * * * * * * * * * * * * *
 *                             *
 *            BUTTONS          *
 *                             *
 * * * * * * * * * * * * * * * */
var buttons = {
  previousState: {}
}

buttons.setLoading = function (name) {
  if (typeof this[name] === 'undefined') return
  let i = this[name].querySelector('i')

  this.previousState[name] = i.innerHTML
  i.style.opacity = 0

  setTimeout(function () {
    i.classList.add('spin')
    i.innerHTML = 'autorenew'
    i.style.opacity = 1
  }, 200)
}

// Changes an element to done animation
buttons.setDone = function (name, success = true) {
  let i = this[name].querySelector('i')

  i.style.opacity = 0

  let thirdStep = () => {
    i.innerHTML = this.previousState[name]
    i.style.opacity = null

    if (selectedItems.length === 0 && document.getElementById('listing')) {
      document.sendCostumEvent('changed-selected')
    }
  }

  let secondStep = () => {
    i.style.opacity = 0
    setTimeout(thirdStep, 200)
  }

  let firstStep = () => {
    i.classList.remove('spin')
    i.innerHTML = success
      ? 'done'
      : 'close'
    i.style.opacity = 1
    setTimeout(secondStep, 1000)
  }

  setTimeout(firstStep, 200)
  return false
}

/* * * * * * * * * * * * * * * *
 *                             *
 *            WEBDAV           *
 *                             *
 * * * * * * * * * * * * * * * */
var webdav = {}

webdav.convertURL = function (url) {
  return window.location.origin + url.replace(data.baseURL + '/', data.webdavURL + '/')
}

webdav.move = function (oldLink, newLink) {
  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    let destination = newLink.replace(data.baseURL + '/', data.webdavURL + '/')

    destination = window.location.origin + destination.substring(data.baseURL.length)

    request.open('MOVE', webdav.convertURL(oldLink), true)
    request.setRequestHeader('Destination', destination)
    request.onload = () => {
      if (request.status === 201 || request.status === 204) {
        resolve()
      } else {
        reject(request.statusText)
      }
    }
    request.onerror = () => reject(request.statusText)
    request.send()
  })
}

webdav.put = function (link, body, headers = {}) {
  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open('PUT', webdav.convertURL(link), true)

    for (let key in headers) {
      request.setRequestHeader(key, headers[key])
    }

    request.onload = () => {
      if (request.status == 201) {
        resolve()
      } else {
        reject(request.statusText)
      }
    }
    request.onerror = () => reject(request.statusText)
    request.send(body)
  })
}

webdav.propfind = function (link, body, headers = {}) {
  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open('PROPFIND', webdav.convertURL(link), true)

    for (let key in headers) {
      request.setRequestHeader(key, headers[key])
    }

    request.onload = () => {
      if (request.status < 300) {
        resolve(request.responseText)
      } else {
        reject(request.statusText)
      }
    }
    request.onerror = () => reject(request.statusText)
    request.send(body)
  })
}

webdav.delete = function (link) {
  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open('DELETE', webdav.convertURL(link), true)
    request.onload = () => {
      if (request.status === 204) {
        resolve()
      } else {
        reject(request.statusText)
      }
    }
    request.onerror = () => reject(request.statusText)
    request.send()
  })
}

webdav.new = function (link) {
  return new Promise((resolve, reject) => {
    let request = new window.XMLHttpRequest()
    request.open((link.endsWith('/') ? 'MKCOL' : 'PUT'), webdav.convertURL(link), true)
    request.onload = () => {
      if (request.status === 201) {
        resolve()
      } else {
        reject(request.statusText)
      }
    }
    request.onerror = () => reject(request.statusText)
    request.send()
  })
}

/* * * * * * * * * * * * * * * *
 *                             *
 *            EVENTS           *
 *                             *
 * * * * * * * * * * * * * * * */
function closePrompt (event) {
  let prompt = document.querySelector('.prompt')

  if (!prompt) return

  if (typeof event !== 'undefined') {
    event.preventDefault()
  }

  document.querySelector('.overlay').classList.remove('active')
  prompt.classList.remove('active')

  setTimeout(() => {
    prompt.remove()
  }, 100)
}

function notImplemented (event) {
  event.preventDefault()
  clickOverlay.click()

  let clone = document.importNode(templates.message.content, true)
  clone.querySelector('h3').innerHTML = 'Not implemented'
  clone.querySelector('p').innerHTML = "Sorry, but this feature wasn't implemented yet."

  document.querySelector('body').appendChild(clone)
  document.querySelector('.overlay').classList.add('active')
  document.querySelector('.prompt').classList.add('active')
}

// Prevent Default event
var preventDefault = function (event) {
  event.preventDefault()
}

function logoutEvent (event) {
  let request = new window.XMLHttpRequest()
  request.open('GET', window.location.pathname, true, 'data.username', 'password')
  request.send()
  request.onreadystatechange = function () {
    if (request.readyState === 4) {
      window.location = '/'
    }
  }
}

function getHash (event, hash) {
  event.preventDefault()

  let request = new window.XMLHttpRequest()
  let link

  if (selectedItems.length) {
    link = document.getElementById(selectedItems[0]).dataset.url
  } else {
    link = window.location.pathname
  }

  request.open('GET', `${link}?checksum=${hash}`, true)

  request.onload = () => {
    if (request.status >= 300) {
      console.log(request.statusText)
      return
    }
    event.target.parentElement.innerHTML = request.responseText
  }
  request.onerror = (e) => console.log(e)
  request.send()
}

function infoEvent (event) {
  event.preventDefault()
  if (event.currentTarget.classList.contains('disabled')) {
    return
  }

  let dir = false
  let link

  if (selectedItems.length) {
    link = document.getElementById(selectedItems[0]).dataset.url
    dir = document.getElementById(selectedItems[0]).dataset.dir
  } else {
    if (document.getElementById('listing') !== null) {
      dir = true
    }

    link = window.location.pathname
  }

  buttons.setLoading('info', false)

  webdav.propfind(link)
    .then((text) => {
      let parser = new window.DOMParser()
      let xml = parser.parseFromString(text, 'text/xml')
      let clone = document.importNode(templates.info.content, true)

      let value = xml.getElementsByTagName('displayname')
      if (value.length > 0) {
        clone.getElementById('display_name').innerHTML = value[0].innerHTML
      } else {
        clone.getElementById('display_name').innerHTML = xml.getElementsByTagName('D:displayname')[0].innerHTML
      }

      value = xml.getElementsByTagName('getcontentlength')
      if (value.length > 0) {
        clone.getElementById('content_length').innerHTML = value[0].innerHTML
      } else {
        clone.getElementById('content_length').innerHTML = xml.getElementsByTagName('D:getcontentlength')[0].innerHTML
      }

      value = xml.getElementsByTagName('getlastmodified')
      if (value.length > 0) {
        clone.getElementById('last_modified').innerHTML = value[0].innerHTML
      } else {
        clone.getElementById('last_modified').innerHTML = xml.getElementsByTagName('D:getlastmodified')[0].innerHTML
      }

      if (dir === true || dir === 'true') {
        clone.querySelector('.file-only').style.display = 'none'
      }

      document.querySelector('body').appendChild(clone)
      document.querySelector('.overlay').classList.add('active')
      document.querySelector('.prompt').classList.add('active')
      buttons.setDone('info', true)
    })
    .catch(e => {
      buttons.setDone('info', false)
      console.log(e)
    })
}

function deleteOnSingleFile () {
  closePrompt()
  buttons.setLoading('delete')

  webdav.delete(window.location.pathname)
    .then(() => {
      window.location.pathname = removeLastDirectoryPartOf(window.location.pathname)
    })
    .catch(e => {
      buttons.setDone('delete', false)
      console.log(e)
    })
}

function deleteOnListing () {
  closePrompt()
  buttons.setLoading('delete')

  let promises = []

  for (let id of selectedItems) {
    promises.push(webdav.delete(document.getElementById(id).dataset.url))
  }

  Promise.all(promises)
    .then(() => {
      listing.reload()
      buttons.setDone('delete')
    })
    .catch(e => {
      console.log(e)
      buttons.setDone('delete', false)
    })
}

// Handles the delete button event
function deleteEvent (event) {
  let single = false

  if (!selectedItems.length) {
    selectedItems = ['placeholder']
    single = true
  }

  let clone = document.importNode(templates.question.content, true)
  clone.querySelector('h3').innerHTML = 'Delete files'

  if (single) {
    clone.querySelector('form').addEventListener('submit', deleteOnSingleFile)
    clone.querySelector('p').innerHTML = `Are you sure you want to delete this file/folder?`
  } else {
    clone.querySelector('form').addEventListener('submit', deleteOnListing)
    clone.querySelector('p').innerHTML = `Are you sure you want to delete ${selectedItems.length} file(s)?`
  }

  clone.querySelector('input').remove()
  clone.querySelector('.ok').innerHTML = 'Delete'

  document.body.appendChild(clone)
  document.querySelector('.overlay').classList.add('active')
  document.querySelector('.prompt').classList.add('active')

  return false
}

function closeHelp (event) {
  event.preventDefault()

  document.querySelector('.help').classList.remove('active')
  document.querySelector('.overlay').classList.remove('active')
}

function openHelp (event) {
  closePrompt(event)

  document.querySelector('.help').classList.add('active')
  document.querySelector('.overlay').classList.add('active')
}

window.addEventListener('keydown', (event) => {
  if (event.keyCode === 27) {
    if (document.querySelector('.help.active')) {
      closeHelp(event)
    }
  }

  if (event.keyCode === 46) {
    deleteEvent(event)
  }

  if (event.keyCode === 112) {
    event.preventDefault()
    openHelp(event)
  }
})

/* * * * * * * * * * * * * * * *
 *                             *
 *           BOOTSTRAP         *
 *                             *
 * * * * * * * * * * * * * * * */

document.addEventListener('DOMContentLoaded', function (event) {
  overlay = document.querySelector('.overlay')
  clickOverlay = document.querySelector('#click-overlay')

  buttons.logout = document.getElementById('logout')
  buttons.delete = document.getElementById('delete')
  buttons.previous = document.getElementById('previous')
  buttons.info = document.getElementById('info')

  // Attach event listeners
  buttons.logout.addEventListener('click', logoutEvent)
  buttons.info.addEventListener('click', infoEvent)

  templates.question = document.querySelector('#question-template')
  templates.info = document.querySelector('#info-template')
  templates.message = document.querySelector('#message-template')
  templates.move = document.querySelector('#move-template')

  if (data.user.AllowEdit) {
    buttons.delete.addEventListener('click', deleteEvent)
  }

  let dropdownButtons = document.querySelectorAll('.action[data-dropdown]')
  Array.from(dropdownButtons).forEach(button => {
    button.addEventListener('click', event => {
      button.querySelector('ul').classList.toggle('active')
      clickOverlay.classList.add('active')

      clickOverlay.addEventListener('click', event => {
        button.querySelector('ul').classList.remove('active')
        clickOverlay.classList.remove('active')
      })
    })
  })

  overlay.addEventListener('click', event => {
    if (document.querySelector('.help.active')) {
      closeHelp(event)
      return
    }

    closePrompt(event)
  })

  let mainActions = document.getElementById('main-actions')

  document.getElementById('more').addEventListener('click', event => {
    event.preventDefault()
    event.stopPropagation()

    clickOverlay.classList.add('active')
    mainActions.classList.add('active')

    clickOverlay.addEventListener('click', event => {
      mainActions.classList.remove('active')
      clickOverlay.classList.remove('active')
    })
  })

  setupSearch()
  return false
})
