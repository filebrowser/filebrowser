import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import { baseURL } from "@/utils/constants";
import { upload as postTus, useTus } from "./tus";
import { createURL, fetchURL, removePrefix, StatusError } from "./utils";
import { isEncodableResponse, makeRawResource } from "@/utils/encodings";

export interface ConvertXTargetsResponse {
  success: boolean;
  from: string;
  targets: Record<string, string[]>;
  message?: string;
}

export interface ConvertXSelection {
  converter: string;
  convertTo: string;
}

export interface ConvertXConvertResponse {
  jobId?: string;
  status: string;
  message?: string;
  source: string;
  destination: string;
  convertTo: string;
  converter?: string;
  done: boolean;
}

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


function wait(ms: number) {
  return new Promise((resolve) => window.setTimeout(resolve, ms));
}

async function pollClamAVScanJob(jobId: string): Promise<ClamAVPathScanResponse> {
  for (;;) {
    await wait(1500);
    const res = await fetchURL(`/api/clamav/jobs/${encodeURIComponent(jobId)}`, {
      method: "GET",
    });
    const result = (await res.json()) as ClamAVPathScanResponse;

    if (result.done || ["clean", "infected", "error"].includes(result.status)) {
      if (result.status === "error") {
        throw new Error(result.message || "ClamAV scan failed");
      }
      return result;
    }
  }
}

async function pollArchiveCreateJob(jobId: string): Promise<ArchiveCreateResponse> {
  for (;;) {
    await wait(1500);
    const res = await fetchURL(`/api/archive/jobs/${encodeURIComponent(jobId)}`, {
      method: "GET",
    });
    const result = (await res.json()) as ArchiveCreateResponse;

    if (result.done || ["done", "error"].includes(result.status || "")) {
      if (result.status === "error") {
        throw new Error(result.message || "Archive creation failed");
      }
      return result;
    }
  }
}

export async function scanWithClamAV(file: ResourceItem): Promise<ClamAVPathScanResponse> {
  const scanPath = removePrefix(file.url);
  const res = await fetchURL(`/api/clamav/scan${scanPath}`, {
    method: "POST",
  });
  const result = (await res.json()) as ClamAVPathScanResponse;

  if (result.jobId && !result.done) {
    return pollClamAVScanJob(result.jobId);
  }

  if (result.status === "error") {
    throw new Error(result.message || "ClamAV scan failed");
  }

  return result;
}

export async function createArchive(
  files: ResourceItem[],
  format = "zip"
): Promise<ArchiveCreateResponse> {
  const res = await fetchURL(`/api/archive/create`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      items: files.map(
        (file) => file.path || decodeURIComponent(removePrefix(file.url))
      ),
      format,
      rename: true,
      overwrite: false,
    }),
  });
  const result = (await res.json()) as ArchiveCreateResponse;

  if (result.jobId && !result.done) {
    return pollArchiveCreateJob(result.jobId);
  }

  if (result.status === "error") {
    throw new Error(result.message || "Archive creation failed");
  }

  return result;
}



async function pollConvertXJob(jobId: string): Promise<ConvertXConvertResponse> {
  for (;;) {
    await wait(1500);
    const res = await fetchURL(`/api/convertx/jobs/${encodeURIComponent(jobId)}`, {
      method: "GET",
    });
    const result = (await res.json()) as ConvertXConvertResponse;

    if (result.done || ["done", "error"].includes(result.status || "")) {
      if (result.status === "error") {
        throw new Error(result.message || "ConvertX conversion failed");
      }
      return result;
    }
  }
}

export async function getConvertXTargets(file: ResourceItem): Promise<ConvertXTargetsResponse> {
  const ext = (file.extension || "").replace(/^\./, "").toLowerCase();
  if (!ext) {
    throw new Error("The selected file has no extension, so ConvertX cannot detect its source format.");
  }

  const params = new URLSearchParams({ from: ext });
  const res = await fetchURL(`/api/convertx/targets?${params}`, {
    method: "GET",
  });
  return (await res.json()) as ConvertXTargetsResponse;
}

export async function convertWithConvertX(
  file: ResourceItem,
  selection: ConvertXSelection
): Promise<ConvertXConvertResponse> {
  const res = await fetchURL(`/api/convertx/convert`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      path: file.path || decodeURIComponent(removePrefix(file.url)),
      convertTo: selection.convertTo,
      converter: selection.converter,
      rename: true,
      overwrite: false,
    }),
  });
  const result = (await res.json()) as ConvertXConvertResponse;

  if (result.jobId && !result.done) {
    return pollConvertXJob(result.jobId);
  }

  if (result.status === "error") {
    throw new Error(result.message || "ConvertX conversion failed");
  }

  return result;
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
