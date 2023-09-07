export default function (name: string) {
  let re = new RegExp(
    "(?:(?:^|.*;\\s*)" + name + "\\s*\\=\\s*([^;]*).*$)|^.*$"
  );
  return document.cookie.replace(re, "$1");
}
