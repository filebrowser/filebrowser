'use strict'

listing.redefineDownloadURLs = function () {
  let files = ''

  for (let i = 0; i < selectedItems.length; i++) {
    let url = document.getElementById(selectedItems[i]).dataset.url
    files += url.replace(window.location.pathname, '') + ','
  }

  files = files.substring(0, files.length - 1)
  files = encodeURIComponent(files)

  let links = document.querySelectorAll('#download ul a')
  Array.from(links).forEach(link => {
    link.href = '?download=' + link.dataset.format + '&files=' + files
  })
}

listing.newFileButton = function (event) {
  event.preventDefault()

  let clone = document.importNode(templates.question.content, true)
  clone.querySelector('h3').innerHTML = 'New file'
  clone.querySelector('p').innerHTML = 'End with a trailing slash to create a dir.'
  clone.querySelector('.ok').innerHTML = 'Create'
  clone.querySelector('form').addEventListener('submit', listing.newFilePrompt)

  document.querySelector('body').appendChild(clone)
  document.querySelector('.overlay').classList.add('active')
  document.querySelector('.prompt').classList.add('active')
}

listing.newFilePrompt = function (event) {
  event.preventDefault()
  buttons.setLoading('new')

  let name = event.currentTarget.querySelector('input').value

  webdav.new(window.location.pathname + name)
    .then(() => {
      buttons.setDone('new')
      listing.reload()
    })
    .catch(e => {
      console.log(e)
      buttons.setDone('new', false)
    })

  closePrompt(event)
  return false
}

listing.addDoubleTapEvent = function () {
  let items = document.getElementsByClassName('item'),
    touches = {
      id: '',
      count: 0
  }

  Array.from(items).forEach(file => {
    file.addEventListener('touchstart', event => {
      if (touches.id != file.id) {
        touches.id = file.id
        touches.count = 1

        setTimeout(() => {
          touches.count = 0
        }, 300)

        return
      }

      touches.count++

      if (touches.count > 1) {
        window.location = file.dataset.url
      }
    })
  })
}


document.addEventListener('DOMContentLoaded', event => {
  listing.addDoubleTapEvent()
  
  if (user.AllowNew) {
    buttons.new.addEventListener('click', listing.newFileButton)    
  }
})
