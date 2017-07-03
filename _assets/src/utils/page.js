function removeLastDir (url) {
  var arr = url.split('/')
  if (arr.pop() === '') {
    arr.pop()
  }

  return arr.join('/')
}

export default {
  removeLastDir: removeLastDir
}
