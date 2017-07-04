function loading (button) {
  let el = document.querySelector(`#${button}-button > i`)

  if (el === undefined || el === null) {
    console.log('Error getting button ' + button)
    return
  }

  el.dataset.icon = el.innerHTML
  el.style.opacity = 0

  setTimeout(() => {
    el.classList.add('spin')
    el.innerHTML = 'autorenew'
    el.style.opacity = 1
  }, 200)
}

function done (button, success = true) {
  let el = document.querySelector(`#${button}-button > i`)

  if (el === undefined || el === null) {
    console.log('Error getting button ' + button)
    return
  }

  el.style.opacity = 0

  let third = () => {
    el.innerHTML = el.dataset.icon
    el.style.opacity = null
  }

  let second = () => {
    el.style.opacity = 0
    setTimeout(third, 200)
  }

  let first = () => {
    el.classList.remove('spin')
    el.innerHTML = success
      ? 'done'
      : 'close'
    el.style.opacity = 1
    setTimeout(second, 200)
  }

  setTimeout(first, 200)
}

export default {
  loading,
  done
}
