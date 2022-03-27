export default function (name) {
  let re = new RegExp(
    "(?:(?:^|.*;\\s*)" + name + "\\s*\\=\\s*([^;]*).*$)|^.*$"
  );
  return document.cookie.replace(re, "$1");
}
