import { fetchURL, removePrefix, createURL } from "./utils";
import { baseURL } from "@/utils/constants";

export async function fetch(url, password = "") {
  url = removePrefix(url);

  const res = await fetchURL(`/api/public/share${url}`, {
    headers: { "X-SHARE-PASSWORD": encodeURIComponent(password) },
  });

  let data = await res.json();
  data.url = `/share${url}`;

  if (data.isDir) {
    if (!data.url.endsWith("/")) data.url += "/";
    data.items = data.items.map((item, index) => {
      item.index = index;
      item.url = `${data.url}${encodeURIComponent(item.name)}`;

      if (item.isDir) {
        item.url += "/";
      }

      return item;
    });
  }

  return data;
}

export function download(format, hash, token, ...files) {
  let url = `${baseURL}/api/public/dl/${hash}`;

  if (files.length === 1) {
    url += encodeURIComponent(files[0]) + "?";
  } else {
    let arg = "";

    for (let file of files) {
      arg += encodeURIComponent(file) + ",";
    }

    arg = arg.substring(0, arg.length - 1);
    arg = encodeURIComponent(arg);
    url += `/?files=${arg}&`;
  }

  if (format) {
    url += `algo=${format}&`;
  }

  if (token) {
    url += `token=${token}&`;
  }

  window.open(url);
}

export function getDownloadURL(share, inline = false) {
  const params = {
    ...(inline && { inline: "true" }),
    ...(share.token && { token: share.token }),
  };

  return createURL("api/public/dl/" + share.hash + share.path, params, false);
}
