// Modified by Lucky Jain (alias: LostB053) on 22/05/2025 (DD/MM/YYYY)
// Modification at line 5, 71-75, 80, 81-83, 84 of this file

import { fetchURL, removePrefix, createURL } from "./utils";
import { baseURL, publicURL } from "@/utils/constants";

export async function fetch(url: string, password: string = "") {
  url = removePrefix(url);

  const res = await fetchURL(
    `/api/public/share${url}`,
    {
      headers: { "X-SHARE-PASSWORD": encodeURIComponent(password) },
    },
    false
  );

  const data = (await res.json()) as Resource;
  data.url = `/share${url}`;

  if (data.isDir) {
    if (!data.url.endsWith("/")) data.url += "/";
    data.items = data.items.map((item: any, index: any) => {
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

export function download(
  format: DownloadFormat,
  hash: string,
  token: string,
  ...files: string[]
) {
  let url = `${baseURL}/api/public/dl/${hash}`;

  if (files.length === 1) {
    url += encodeURIComponent(files[0]) + "?";
  } else {
    let arg = "";

    for (const file of files) {
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

export function getDownloadURL(
  res: Resource,
  inline = false,
  singleShare = false
) {
  const params = {
    ...(inline && { inline: "true" }),
    ...(res.token && { token: res.token }),
  };
  const prefix = publicURL ? "" : "api/public/dl/";
  const filename = singleShare
    ? "/" + res.path.match(/\/([^\/]+)$/)?.[1]
    : res.path;
  return createURL(prefix + res.hash + filename, params, true, true);
}
