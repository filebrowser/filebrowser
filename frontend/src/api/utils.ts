import { useAuthStore } from "@/stores/auth";
import { renew, logout } from "@/utils/auth";
import { baseURL } from "@/utils/constants";
import { encodePath } from "@/utils/url";

export class StatusError extends Error {
  constructor(
    message: any,
    public status?: number,
    public is_canceled?: boolean
  ) {
    super(message);
    this.name = "StatusError";
  }
}

export async function fetchURL(
  url: string,
  opts: ApiOpts,
  auth = true
): Promise<Response> {
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
  } catch (e) {
    // Check if the error is an intentional cancellation
    if (e instanceof Error && e.name === "AbortError") {
      throw new StatusError("000 No connection", 0, true);
    }
    throw new StatusError("000 No connection", 0);
  }

  if (auth && res.headers.get("X-Renew-Token") === "true") {
    await renew(authStore.jwt);
  }

  if (res.status < 200 || res.status > 299) {
    const body = await res.text();
    const error = new StatusError(
      body || `${res.status} ${res.statusText}`,
      res.status
    );

    if (auth && res.status == 401) {
      logout();
    }

    throw error;
  }

  return res;
}

export async function fetchJSON<T>(url: string, opts?: any): Promise<T> {
  const res = await fetchURL(url, opts);

  if (res.status === 200) {
    return res.json() as Promise<T>;
  }

  throw new StatusError(`${res.status} ${res.statusText}`, res.status);
}

export function removePrefix(url: string): string {
  url = url.split("/").splice(2).join("/");

  if (url === "") url = "/";
  if (url[0] !== "/") url = "/" + url;
  return url;
}

export function createURL(endpoint: string, searchParams = {}): string {
  let prefix = baseURL;
  if (!prefix.endsWith("/")) {
    prefix = prefix + "/";
  }
  const url = new URL(prefix + encodePath(endpoint), origin);
  url.search = new URLSearchParams(searchParams).toString();

  return url.toString();
}

export function setSafeTimeout(callback: () => void, delay: number): number {
  const MAX_DELAY = 86_400_000;
  let remaining = delay;

  function scheduleNext(): number {
    if (remaining <= MAX_DELAY) {
      return window.setTimeout(callback, remaining);
    } else {
      return window.setTimeout(() => {
        remaining -= MAX_DELAY;
        scheduleNext();
      }, MAX_DELAY);
    }
  }

  return scheduleNext();
}

export function clearUploadList(
  files: UploadList,
  result: Array<ConflictingResource>
) {
  for (let i = result.length - 1; i >= 0; i--) {
    const item = result[i];

    // Rename (In upload files is no supported rename yet, it will cause 409 error)
    if (item.checked.length == 2) {
      continue;
    }

    const base = item.name.replace(/^\/?[^/]+\//, "");
    // Overwrite (Mark as overwrite file/folder and subfolders from list)
    if (item.checked.length == 1 && item.checked[0] === "origin") {
      for (const f of files) {
        const fullPath = f.fullPath ?? f.name;
        if (fullPath === base || fullPath.startsWith(base + "/")) {
          f.overwrite = true;
        }
      }
      continue;
    }

    // Skip (delete file/folder and subfolders from list)
    for (let i = files.length - 1; i >= 0; i--) {
      const f = files[i];
      const fullPath = f.fullPath ?? f.name;
      if (fullPath === base || fullPath.startsWith(base + "/")) {
        files.splice(i, 1);
      }
    }
  }
  return files;
}

export function clearCopyMoveList(
  files: Array<any>,
  result: Array<ConflictingResource>
) {
  for (let i = result.length - 1; i >= 0; i--) {
    const item = result[i];
    const base = item.name;
    const prefix = "/files" + base;

    // Rename (Mark as rename file/folder and subfolders from list)
    if (item.checked.length === 2) {
      for (const f of files) {
        if (f.to === prefix || f.to.startsWith(prefix + "/")) {
          f.rename = true;
        }
      }
      continue;
    }

    // Overwrite (Mark as overwrite file/folder and subfolders from list)
    if (item.checked.length === 1 && item.checked[0] === "origin") {
      for (const f of files) {
        if (f.to === prefix || f.to.startsWith(prefix + "/")) {
          f.overwrite = true;
        }
      }
      continue;
    }

    // Skip (delete file/folder and subfolders from list)
    for (let j = files.length - 1; j >= 0; j--) {
      const f = files[j];
      if (f.to === prefix || f.to.startsWith(prefix + "/")) {
        files.splice(j, 1);
      }
    }
  }

  console.log(files);
  return files;
}
