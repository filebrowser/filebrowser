// Removes an element, if exists, from an array
function removeElement (array, element) {
  var i = array.indexOf(element)
  if (i !== -1) {
    array.splice(i, 1)
  }

  return array
}

// Replaces an element inside an array by another
function replaceElement (array, oldElement, newElement) {
  var i = array.indexOf(oldElement)
  if (i !== -1) {
    array[i] = newElement
  }

  return array
}

export default {
  remove: removeElement,
  replace: replaceElement
}
