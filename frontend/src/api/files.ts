import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import { baseURL } from "@/utils/constants";
import { upload as postTus, useTus } from "./tus";
import { createURL, fetchURL, removePrefix, StatusError } from "./utils";
import { isEncodableResponse, makeRawResource } from "@/utils/encodings";

export async function fetch(url: string, signal?: AbortSignal) {
  const encoding = isEncodableResponse(url);
  url = removePrefix(url);
  const res = await fetchURL(`/api/resources${url}`, {
    signal,
    headers: {
      "X-Encoding": encoding ? "true" : "false",
    },
  });

  let data: Resource;
  try {
    if (res.headers.get("Content-Type") == "application/octet-stream") {
      data = await makeRawResource(res, url);
    } else {
      data = (await res.json()) as Resource;
    }
  } catch (e) {
    // Check if the error is an intentional cancellation
    if (e instanceof Error && e.name === "AbortError") {
      throw new StatusError("000 No connection", 0, true);
    }
    throw e;
  }
  data.url = `/files${url}`;

  if (data.isDir) {
    if (!data.url.endsWith("/")) data.url += "/";
    // Perhaps change the any
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


function archiveRoute(archivePath: string, innerPath: string) {
  const inner = innerPath || "/";
  return `${archivePath}?archive=${encodeURIComponent(inner)}`;
}

function normalizeArchiveInner(innerPath: string) {
  if (!innerPath || innerPath === ".") return "/";
  if (!innerPath.startsWith("/")) return `/${innerPath}`;
  return innerPath;
}

export async function fetchArchive(
  archiveURL: string,
  innerPath = "/",
  signal?: AbortSignal
) {
  const archivePath = removePrefix(archiveURL);
  const archiveViewPath = `/files${archivePath}`;
  const inner = normalizeArchiveInner(innerPath);
  const params = new URLSearchParams({ inner });
  const res = await fetchURL(`/api/archive/resources${archivePath}?${params}`, {
    signal,
  });

  let data: Resource;
  try {
    data = (await res.json()) as Resource;
  } catch (e) {
    if (e instanceof Error && e.name === "AbortError") {
      throw new StatusError("000 No connection", 0, true);
    }
    throw e;
  }

  data.archive = true;
  data.archivePath = archiveViewPath;
  data.archiveInnerPath = normalizeArchiveInner(data.archiveInnerPath || inner);
  data.url = archiveRoute(archiveViewPath, data.archiveInnerPath);

  if (data.isDir) {
    data.items = data.items.map((item: any, index: any) => {
      item.index = index;
      item.archive = true;
      item.archivePath = archiveViewPath;
      item.archiveInnerPath = normalizeArchiveInner(
        item.archiveInnerPath || item.path
      );
      item.url = archiveRoute(archiveViewPath, item.archiveInnerPath);
      return item;
    });
  }

  return data;
}

export async function fetchAll(url: string): Promise<RecursiveEntry[]> {
  url = removePrefix(url);
  const res = await fetchURL(`/api/resources/recursive${url}`, {});
  return (await res.json()) as RecursiveEntry[];
}

async function resourceAction(url: string, method: ApiMethod, content?: any) {
  url = removePrefix(url);

  const opts: ApiOpts = {
    method,
  };

  if (content) {
    opts.body = content;
  }

  const res = await fetchURL(`/api/resources${url}`, opts);

  return res;
}

export async function remove(url: string) {
  return resourceAction(url, "DELETE");
}

export async function put(url: string, content = "") {
  return resourceAction(url, "PUT", content);
}

export function download(format: any, ...files: string[]) {
  let url = `${baseURL}/api/raw`;

  if (files.length === 1) {
    url += removePrefix(files[0]) + "?";
  } else {
    let arg = "";

    for (const file of files) {
      arg += removePrefix(file) + ",";
    }

    arg = arg.substring(0, arg.length - 1);
    arg = encodeURIComponent(arg);
    url += `/?files=${arg}&`;
  }

  if (format) {
    url += `algo=${format}&`;
  }

  window.open(url);
}

export async function post(
  url: string,
  content: ApiContent = "",
  overwrite = false,
  onupload: any = () => {}
) {
  // Use the pre-existing API if:
  const useResourcesApi =
    // a folder is being created
    url.endsWith("/") ||
    // We're not using http(s)
    (content instanceof Blob &&
      !["http:", "https:"].includes(window.location.protocol)) ||
    // Tus is disabled / not applicable
    !(await useTus(content));
  return useResourcesApi
    ? postResources(url, content, overwrite, onupload)
    : postTus(url, content, overwrite, onupload);
}

async function postResources(
  url: string,
  content: ApiContent = "",
  overwrite = false,
  onupload: any
) {
  url = removePrefix(url);

  let bufferContent: ArrayBuffer;
  if (
    content instanceof Blob &&
    !["http:", "https:"].includes(window.location.protocol)
  ) {
    bufferContent = await new Response(content).arrayBuffer();
  }

  const authStore = useAuthStore();
  return new Promise((resolve, reject) => {
    const request = new XMLHttpRequest();
    request.open(
      "POST",
      `${baseURL}/api/resources${url}?override=${overwrite}`,
      true
    );
    request.setRequestHeader("X-Auth", authStore.jwt);

    if (typeof onupload === "function") {
      request.upload.onprogress = onupload;
    }

    request.onload = () => {
      if (request.status === 200) {
        resolve(request.responseText);
      } else if (request.status === 409) {
        reject(new Error(request.status.toString()));
      } else {
        reject(new Error(request.responseText));
      }
    };

    request.onerror = () => {
      reject(new Error("001 Connection aborted"));
    };

    request.send(bufferContent || content);
  });
}

function moveCopy(
  items: any[],
  copy = false,
  overwrite = false,
  rename = false
) {
  const layoutStore = useLayoutStore();
  const promises = [];

  for (const item of items) {
    const from = item.from;
    const to = encodeURIComponent(removePrefix(item.to ?? ""));
    const finalOverwrite =
      item.overwrite == undefined ? overwrite : item.overwrite;
    const finalRename = item.rename == undefined ? rename : item.rename;
    const url = `${from}?action=${
      copy ? "copy" : "rename"
    }&destination=${to}&override=${finalOverwrite}&rename=${finalRename}`;
    promises.push(resourceAction(url, "PATCH"));
  }
  layoutStore.closeHovers();
  return Promise.all(promises);
}

export function move(items: any[], overwrite = false, rename = false) {
  return moveCopy(items, false, overwrite, rename);
}

export function copy(items: any[], overwrite = false, rename = false) {
  return moveCopy(items, true, overwrite, rename);
}

export async function checksum(url: string, algo: ChecksumAlg) {
  const data = await resourceAction(`${url}?checksum=${algo}`, "GET");
  return (await data.json()).checksums[algo];
}

export function getDownloadURL(file: ResourceItem, inline: any) {
  const params: Record<string, string> = {
    ...(inline && { inline: "true" }),
  };

  if (file.archivePath && file.archiveInnerPath) {
    params.inner = file.archiveInnerPath;
    return createURL("api/archive/raw" + removePrefix(file.archivePath), params);
  }

  return createURL("api/raw" + file.path, params);
}

export function getPreviewURL(file: ResourceItem, size: string) {
  if (file.archivePath && file.archiveInnerPath) {
    return getDownloadURL(file, true);
  }

  const params = {
    inline: "true",
    key: Date.parse(file.modified),
  };

  return createURL("api/preview/" + size + file.path, params);
}


export async function extractArchive(
  archiveURL: string,
  innerPath = "/",
  destination = "",
  rename = true,
  overwrite = false
): Promise<{ destination: string; extracted: number }> {
  const archivePath = removePrefix(archiveURL);
  const params = new URLSearchParams({
    inner: normalizeArchiveInner(innerPath),
    rename: String(rename),
    override: String(overwrite),
  });
  if (destination) {
    params.set("destination", removePrefix(destination));
  }

  const res = await fetchURL(`/api/archive/extract${archivePath}?${params}`, {
    method: "POST",
  });
  return (await res.json()) as { destination: string; extracted: number };
}

export function getSubtitlesURL(file: ResourceItem) {
  const params = {
    inline: "true",
  };

  return file.subtitles?.map((d) => createURL("api/subtitle" + d, params));
}

export async function usage(url: string, signal: AbortSignal) {
  url = removePrefix(url);

  const res = await fetchURL(`/api/usage${url}`, { signal });

  try {
    return await res.json();
  } catch (e) {
    // Check if the error is an intentional cancellation
    if (e instanceof Error && e.name == "AbortError") {
      throw new StatusError("000 No connection", 0, true);
    }
    throw e;
  }
}
