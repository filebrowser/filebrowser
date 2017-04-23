'use strict'

var listing = {
  selectMultiple: false
}

listing.reload = function (callback) {
  let request = new XMLHttpRequest()

  request.open('GET', window.location)
  request.setRequestHeader('Minimal', 'true')
  request.send()
  request.onreadystatechange = function () {
    if (request.readyState === 4) {
      if (request.status === 200) {
        document.querySelector('body main').innerHTML = request.responseText
        listing.addDoubleTapEvent()

        if (typeof callback === 'function') {
          callback()
        }
      }
    }
  }
}

listing.itemDragStart = function (event) {
  let el = event.target

  for (let i = 0; i < 5; i++) {
    if (!el.classList.contains('item')) {
      el = el.parentElement
    }
  }

  event.dataTransfer.setData('id', el.id)
  event.dataTransfer.setData('name', el.querySelector('.name').innerHTML)
}

listing.itemDragOver = function (event) {
  event.preventDefault()
  let el = event.target

  for (let i = 0; i < 5; i++) {
    if (!el.classList.contains('item')) {
      el = el.parentElement
    }
  }

  el.style.opacity = 1
}

listing.itemDrop = function (e) {
  e.preventDefault()

  let el = e.target,
    id = e.dataTransfer.getData('id'),
    name = e.dataTransfer.getData('name')

  if (id == '' || name == '') return

  for (let i = 0; i < 5; i++) {
    if (!el.classList.contains('item')) {
      el = el.parentElement
    }
  }

  if (el.id === id) return

  let oldLink = document.getElementById(id).dataset.url,
    newLink = el.dataset.url + name

  webdav.move(oldLink, newLink)
    .then(() => listing.reload())
    .catch(e => console.log(e))
}

listing.documentDrop = function (event) {
  event.preventDefault()
  let dt = event.dataTransfer,
    files = dt.files,
    el = event.target,
    items = document.getElementsByClassName('item')

  for (let i = 0; i < 5; i++) {
    if (el != null && !el.classList.contains('item')) {
      el = el.parentElement
    }
  }

  if (files.length > 0) {
    if (el != null && el.classList.contains('item') && el.dataset.dir == 'true') {
      listing.handleFiles(files, el.querySelector('.name').innerHTML + '/')
      return
    }

    listing.handleFiles(files, '')
  } else {
    Array.from(items).forEach(file => {
      file.style.opacity = 1
    })
  }
}

listing.rename = function (event) {
  if (!selectedItems.length || selectedItems.length > 1) {
    return false
  }

  let item = document.getElementById(selectedItems[0])

  if (item.classList.contains('disabled')) {
    return false
  }

  let link = item.dataset.url,
    field = item.querySelector('.name'),
    name = field.innerHTML

  let submit = (event) => {
    event.preventDefault()

    let newName = event.currentTarget.querySelector('input').value,
      newLink = removeLastDirectoryPartOf(link) + '/' + newName

    closePrompt(event)
    buttons.setLoading('rename')

    webdav.move(link, newLink).then(() => {
      listing.reload(() => {
        newName = btoa(newName)
        selectedItems = [newName]
        document.getElementById(newName).setAttribute('aria-selected', true)
        listing.handleSelectionChange()
      })

      buttons.setDone('rename')
    }).catch(error => {
      field.innerHTML = name
      buttons.setDone('rename', false)
      console.log(error)
    })

    return false
  }

  let clone = document.importNode(templates.question.content, true)
  clone.querySelector('h3').innerHTML = 'Rename'
  clone.querySelector('input').value = name
  clone.querySelector('.ok').innerHTML = 'Rename'
  clone.querySelector('form').addEventListener('submit', submit)

  document.querySelector('body').appendChild(clone)
  document.querySelector('.overlay').classList.add('active')
  document.querySelector('.prompt').classList.add('active')

  return false
}

listing.handleFiles = function (files, base) {
  buttons.setLoading('upload')

  let promises = []

  for (let file of files) {
    promises.push(webdav.put(window.location.pathname + base + file.name, file))
  }

  Promise.all(promises)
    .then(() => {
      listing.reload()
      buttons.setDone('upload')
    })
    .catch(e => {
      console.log(e)
      buttons.setDone('upload', false)
    })

  return false
}

listing.unselectAll = function () {
  let items = document.getElementsByClassName('item')
  Array.from(items).forEach(link => {
    link.setAttribute('aria-selected', false)
  })

  selectedItems = []

  listing.handleSelectionChange()
  return false
}

listing.handleSelectionChange = function (event) {
  listing.redefineDownloadURLs()

  let selectedNumber = selectedItems.length,
    fileAction = document.getElementById('file-only')

  if (selectedNumber) {
    fileAction.classList.remove('disabled')

    if (selectedNumber > 1) {
      buttons.open.classList.add('disabled')
      buttons.rename.classList.add('disabled')
      buttons.info.classList.add('disabled')
    }

    if (selectedNumber == 1) {
      if (document.getElementById(selectedItems[0]).dataset.dir == 'true') {
        buttons.open.classList.add('disabled')
      } else {
        buttons.open.classList.remove('disabled')
      }

      buttons.info.classList.remove('disabled')
      buttons.rename.classList.remove('disabled')
    }

    return false
  }

  fileAction.classList.add('disabled')
  return false
}

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

listing.openItem = function (event) {
  window.location = event.currentTarget.dataset.url
}

listing.selectItem = function (event) {
  let el = event.currentTarget

  if (selectedItems.length != 0) event.preventDefault()
  if (selectedItems.indexOf(el.id) == -1) {
    if (!event.ctrlKey && !listing.selectMultiple) listing.unselectAll()

    el.setAttribute('aria-selected', true)
    selectedItems.push(el.id)
  } else {
    el.setAttribute('aria-selected', false)
    selectedItems.removeElement(el.id)
  }

  listing.handleSelectionChange()
  return false
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

listing.updateColumns = function (event) {
  let columns = Math.floor(document.getElementById('listing').offsetWidth / 300),
    items = getCSSRule(['#listing.mosaic .item', '.mosaic#listing .item'])

  items.style.width = `calc(${100/columns}% - 1em)`
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

// Keydown events
window.addEventListener('keydown', (event) => {
  if (event.keyCode == 27) {
    listing.unselectAll()

    if (document.querySelectorAll('.prompt').length) {
      closePrompt(event)
    }
  }

  if (event.keyCode == 113) {
    listing.rename()
  }

  if (event.ctrlKey || event.metaKey) {
    switch (String.fromCharCode(event.which).toLowerCase()) {
      case 's':
        event.preventDefault()
        window.location = '?download=true'
    }
  }
})

window.addEventListener('resize', () => {
  listing.updateColumns()
})

listing.selectMoveFolder = function (event) {
  if (event.target.getAttribute('aria-selected') === 'true') {
    event.target.setAttribute('aria-selected', false)
    return
  } else {
    if (document.querySelector('.file-list li[aria-selected=true]')) {
      document.querySelector('.file-list li[aria-selected=true]').setAttribute('aria-selected', false)
    }
    event.target.setAttribute('aria-selected', true)
    return
  }
}

listing.getJSON = function (link) {
  return new Promise((resolve, reject) => {
    let request = new XMLHttpRequest()
    request.open('GET', link)
    request.setRequestHeader('Accept', 'application/json')
    request.onload = () => {
      if (request.status == 200) {
        resolve(request.responseText)
      } else {
        reject(request.statusText)
      }
    }
    request.onerror = () => reject(request.statusText)
    request.send()
  })
}

listing.moveMakeItem = function (url, name) {
  let node = document.createElement('li'),
    count = 0

  node.dataset.url = url
  node.innerHTML = name
  node.setAttribute('aria-selected', false)

  node.addEventListener('dblclick', listing.moveDialogNext)
  node.addEventListener('click', listing.selectMoveFolder)
  node.addEventListener('touchstart', event => {
    count++

    setTimeout(() => {
      count = 0
    }, 300)

    if (count > 1) {
      listing.moveDialogNext(event)
    }
  })

  return node
}

listing.moveDialogNext = function (event) {
  let request = new XMLHttpRequest(),
    prompt = document.querySelector('form.prompt.active'),
    list = prompt.querySelector('div.file-list ul')

  prompt.addEventListener('submit', listing.moveSelected)

  listing.getJSON(event.target.dataset.url)
    .then((data) => {
      let dirs = 0

      prompt.querySelector('ul').innerHTML = ''
      prompt.querySelector('code').innerHTML = event.target.dataset.url

      if (event.target.dataset.url != baseURL + '/') {
        let node = listing.moveMakeItem(removeLastDirectoryPartOf(event.target.dataset.url) + '/', '..')
        list.appendChild(node)
      }

      if (JSON.parse(data) == null) {
        prompt.querySelector('p').innerHTML = `There aren't any folders in this directory.`
        return
      }

      for (let f of JSON.parse(data)) {
        if (f.IsDir === true) {
          dirs++
          list.appendChild(listing.moveMakeItem(f.URL, f.Name))
        }
      }

      if (dirs === 0)
        prompt.querySelector('p').innerHTML = `There aren't any folders in this directory.`
    })
    .catch(e => console.log(e))
}

listing.moveSelected = function (event) {
  event.preventDefault()

  let promises = []
  buttons.setLoading('move')

  for (let file of selectedItems) {
    let fileElement = document.getElementById(file),
      destFolder = event.target.querySelector('p code').innerHTML

    if (event.currentTarget.querySelector('li[aria-selected=true]') != null) {
      destFolder = event.currentTarget.querySelector('li[aria-selected=true]').dataset.url
    }

    let destPath = '/' + destFolder + '/' + fileElement.querySelector('.name').innerHTML
    destPath = destPath.replace('//', '/')

    promises.push(webdav.move(fileElement.dataset.url, destPath))
  }

  Promise.all(promises)
    .then(() => {
      closePrompt(event)
      buttons.setDone('move')
      listing.reload()
    })
    .catch(e => {
      console.log(e)
    })
}

listing.moveEvent = function (event) {
  if (event.currentTarget.classList.contains('disabled'))
    return

  listing.getJSON(window.location.pathname)
    .then((data) => {
      let prompt = document.importNode(templates.move.content, true),
        list = prompt.querySelector('div.file-list ul'),
        dirs = 0

      prompt.querySelector('form').addEventListener('submit', listing.moveSelected)
      prompt.querySelector('code').innerHTML = window.location.pathname

      if (window.location.pathname !== baseURL + '/') {
        list.appendChild(listing.moveMakeItem(removeLastDirectoryPartOf(window.location.pathname) + '/', '..'))
      }

      for (let f of JSON.parse(data)) {
        if (f.IsDir === true) {
          dirs++
          list.appendChild(listing.moveMakeItem(f.URL, f.Name))
        }
      }

      if (dirs === 0) {
        prompt.querySelector('p').innerHTML = `There aren't any folders in this directory.`
      }

      document.body.appendChild(prompt)
      document.querySelector('.overlay').classList.add('active')
      document.querySelector('.prompt').classList.add('active')
    })
    .catch(e => console.log(e))
}

document.addEventListener('DOMContentLoaded', event => {
  listing.updateColumns()
  listing.addDoubleTapEvent()

  buttons.rename = document.getElementById('rename')
  buttons.upload = document.getElementById('upload')
  buttons.new = document.getElementById('new')
  buttons.download = document.getElementById('download')
  buttons.move = document.getElementById('move')

  buttons.move.addEventListener('click', listing.moveEvent)

  document.getElementById('multiple-selection-activate').addEventListener('click', event => {
    listing.selectMultiple = true
    clickOverlay.click()

    document.getElementById('multiple-selection').classList.add('active')
    document.querySelector('body').style.paddingBottom = '4em'
  })

  document.getElementById('multiple-selection-cancel').addEventListener('click', event => {
    listing.selectMultiple = false

    document.querySelector('body').style.paddingBottom = '0'
    document.getElementById('multiple-selection').classList.remove('active')
  })

  if (user.AllowEdit) {
    buttons.rename.addEventListener('click', listing.rename)
  }

  let items = document.getElementsByClassName('item')

  if (user.AllowNew) {
    buttons.upload.addEventListener('click', (event) => {
      document.getElementById('upload-input').click()
    })

    buttons.new.addEventListener('click', listing.newFileButton)

    // Drag and Drop
    document.addEventListener('dragover', function (event) {
      event.preventDefault()
    }, false)

    document.addEventListener('dragenter', (event) => {
      Array.from(items).forEach(file => {
        file.style.opacity = 0.5
      })
    }, false)

    document.addEventListener('dragend', (event) => {
      Array.from(items).forEach(file => {
        file.style.opacity = 1
      })
    }, false)

    document.addEventListener('drop', listing.documentDrop, false)
  }
})
