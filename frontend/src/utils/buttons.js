function loading(button) {
  const el = document.querySelector(`#${button}-button > i`)

  if (el === undefined || el === null) {
    console.log('Error getting button ' + button) // eslint-disable-line
    return
  }

  el.dataset.icon = el.innerHTML
  el.style.opacity = 0

  setTimeout(() => {
    el.classList.add('spin')
    el.innerHTML = 'autorenew'
    el.style.opacity = 1
  }, 100)
}

function done(button) {
  const el = document.querySelector(`#${button}-button > i`)

  if (el === undefined || el === null) {
    console.log('Error getting button ' + button) // eslint-disable-line
    return
  }

  el.style.opacity = 0

  setTimeout(() => {
    el.classList.remove('spin')
    el.innerHTML = el.dataset.icon
    el.style.opacity = 1
  }, 100)
}

function success(button) {
  const el = document.querySelector(`#${button}-button > i`)

  if (el === undefined || el === null) {
    console.log('Error getting button ' + button) // eslint-disable-line
    return
  }

  el.style.opacity = 0

  setTimeout(() => {
    el.classList.remove('spin')
    el.innerHTML = 'done'
    el.style.opacity = 1

    setTimeout(() => {
      el.style.opacity = 0

      setTimeout(() => {
        el.innerHTML = el.dataset.icon
        el.style.opacity = 1
      }, 100)
    }, 500)
  }, 100)
}

export default {
  loading,
  done,
  success
}
