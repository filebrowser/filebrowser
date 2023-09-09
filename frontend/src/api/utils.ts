import { useAuthStore } from "@/stores/auth";
import { renew, logout } from "@/utils/auth";
import { baseURL } from "@/utils/constants";
import { encodePath } from "@/utils/url";

export async function fetchURL(url: ApiUrl, opts: ApiOpts, auth = true) {
  const authStore = useAuthStore();

  opts = opts || {};
  opts.headers = opts.headers || {};

  let { headers, ...rest } = opts;
  let res;
  try {
    res = await fetch(`${baseURL}${url}`, {
      headers: {
        "X-Auth": authStore.jwt,
        ...headers,
      },
      ...rest,
    });
  } catch {
    const error = new Error("000 No connection");
    // @ts-ignore don't know yet how to solve
    error.status = 0;

    throw error;
  }

  if (auth && res.headers.get("X-Renew-Token") === "true") {
    await renew(authStore.jwt);
  }

  if (res.status < 200 || res.status > 299) {
    const error = new Error(await res.text());
    // @ts-ignore don't know yet how to solve
    error.status = res.status;

    if (auth && res.status == 401) {
      logout();
    }

    throw error;
  }

  return res;
}

export async function fetchJSON(url: ApiUrl, opts?: any) {
  const res = await fetchURL(url, opts);

  if (res.status === 200) {
    return res.json();
  } else {
    throw new Error(res.status.toString());
  }
}

export function removePrefix(url: ApiUrl) {
  url = url.split("/").splice(2).join("/");

  if (url === "") url = "/";
  if (url[0] !== "/") url = "/" + url;
  return url;
}

export function createURL(endpoint: ApiUrl, params = {}, auth = true) {
  const authStore = useAuthStore();

  let prefix = baseURL;
  if (!prefix.endsWith("/")) {
    prefix = prefix + "/";
  }
  const url = new URL(prefix + encodePath(endpoint), origin);

  const searchParams: searchParams = {
    ...(auth && { auth: authStore.jwt }),
    ...params,
  };

  for (const key in searchParams) {
    url.searchParams.set(key, searchParams[key]);
  }

  return url.toString();
}
