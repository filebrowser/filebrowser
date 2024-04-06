import { fetchURL, removePrefix, createURL } from "./utils";
import { baseURL } from "@/utils/constants";
import { useAuthStore } from "@/stores/auth";
export async function fetch(url: string, password: string = "") {
  url = removePrefix(url);

  const authStore = useAuthStore();
  const res = await fetchURL(
    `/api/public/share${url}?s=${authStore.shareConfig.sortBy}&a=${Number(authStore.shareConfig.asc)}`,
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

export function getDownloadURL(res: Resource, inline = false) {
  const params = {
    ...(inline && { inline: "true" }),
    ...(res.token && { token: res.token }),
  };

  return createURL("api/public/dl/" + res.hash + res.path, params, false);
}

export function getPreviewURL(file: ResourceItem, size: string) {
  const authStore = useAuthStore();
  const params = {
    inline: "true",
    token: authStore.guestJwt,
    key: Date.parse(file.modified),
  };

  return createURL("api/public/preview/" + size + file.path, params, false);
}
