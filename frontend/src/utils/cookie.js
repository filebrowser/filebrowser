export default function(name) {
  const re = new RegExp('(?:(?:^|.*;\\s*)' + name + '\\s*\\=\\s*([^;]*).*$)|^.*$')
  return document.cookie.replace(re, '$1')
}
