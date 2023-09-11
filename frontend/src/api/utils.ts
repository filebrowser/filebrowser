import { useAuthStore } from "@/stores/auth";
import type { ApiOpts, SearchParams } from "@/types";
import { renew, logout } from "@/utils/auth";
import { baseURL } from "@/utils/constants";
import { encodePath } from "@/utils/url";

export class StatusError extends Error {
  constructor(
    message: any,
    public status?: number
  ) {
    super(message);
    this.name = "StatusError";
  }
}

export async function fetchURL(url: string, opts: ApiOpts, auth = true) {
  const authStore = useAuthStore();

  opts = opts || {};
  opts.headers = opts.headers || {};

  const { headers, ...rest } = opts;
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
    throw new StatusError("000 No connection", 0);
  }

  if (auth && res.headers.get("X-Renew-Token") === "true") {
    await renew(authStore.jwt);
  }

  if (res.status < 200 || res.status > 299) {
    const error = new StatusError(await res.text(), res.status);

    if (auth && res.status == 401) {
      logout();
    }

    throw error;
  }

  return res;
}

export async function fetchJSON(url: string, opts?: any) {
  const res = await fetchURL(url, opts);

  if (res.status === 200) {
    return res.json();
  } else {
    throw new Error(res.status.toString());
  }
}

export function removePrefix(url: string) {
  url = url.split("/").splice(2).join("/");

  if (url === "") url = "/";
  if (url[0] !== "/") url = "/" + url;
  return url;
}

export function createURL(endpoint: string, params = {}, auth = true) {
  const authStore = useAuthStore();

  let prefix = baseURL;
  if (!prefix.endsWith("/")) {
    prefix = prefix + "/";
  }
  const url = new URL(prefix + encodePath(endpoint), origin);

  const searchParams: SearchParams = {
    ...(auth && { auth: authStore.jwt }),
    ...params,
  };

  for (const key in searchParams) {
    url.searchParams.set(key, searchParams[key]);
  }

  return url.toString();
}
