import store from "@/store";
import { renew } from "@/utils/auth";
import { baseURL } from "@/utils/constants";
import { encodePath } from "@/utils/url";

export async function fetchURL(url, opts) {
  opts = opts || {};
  opts.headers = opts.headers || {};

  let { headers, ...rest } = opts;

  let res;
  try {
    res = await fetch(`${baseURL}${url}`, {
      headers: {
        "X-Auth": store.state.jwt,
        ...headers,
      },
      ...rest,
    });
  } catch (error) {
    return { status: 0 };
  }

  if (res.headers.get("X-Renew-Token") === "true") {
    await renew(store.state.jwt);
  }

  return res;
}

export async function fetchJSON(url, opts) {
  const res = await fetchURL(url, opts);

  if (res.status === 200) {
    return res.json();
  } else {
    throw new Error(res.status);
  }
}

export function removePrefix(url) {
  url = url.split("/").splice(2).join("/");

  if (url === "") url = "/";
  if (url[0] !== "/") url = "/" + url;
  return url;
}

export function createURL(endpoint, params = {}, auth = true) {
  const url = new URL(encodePath(endpoint), origin + baseURL);

  const searchParams = {
    ...(auth && { auth: store.state.jwt }),
    ...params,
  };

  for (const key in searchParams) {
    url.searchParams.set(key, searchParams[key]);
  }

  return url.toString();
}
