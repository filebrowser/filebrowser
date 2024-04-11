export default function (name: string) {
  const re = new RegExp(
    "(?:(?:^|.*;\\s*)" + name + "\\s*\\=\\s*([^;]*).*$)|^.*$"
  );
  return document.cookie.replace(re, "$1");
}
