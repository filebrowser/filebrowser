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
