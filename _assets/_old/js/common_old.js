'use strict'

var tempID = '_fm_internal_temporary_id'

var templates = {}
var selectedItems = []
var overlay
var clickOverlay

// Sends a costum event to itself
Document.prototype.sendCostumEvent = function (text) {
  this.dispatchEvent(new window.CustomEvent(text))
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
