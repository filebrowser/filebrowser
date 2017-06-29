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
