import { createURL, fetchURL, removePrefix } from "./utils";
import { baseURL } from "@/utils/constants";
import store from "@/store";
import { upload as postTus, useTus } from "./tus";

export async function fetch(url) {
  url = removePrefix(url);

  const res = await fetchURL(`/api/resources${url}`, {});

  let data = await res.json();
  data.url = `/files${url}`;

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

async function resourceAction(url, method, content) {
  url = removePrefix(url);

  let opts = { method };

  if (content) {
    opts.body = content;
  }

  const res = await fetchURL(`/api/resources${url}`, opts);

  return res;
}

export async function remove(url) {
  return resourceAction(url, "DELETE");
}

export async function put(url, content = "") {
  return resourceAction(url, "PUT", content);
}

export function download(format, ...files) {
  let url = `${baseURL}/api/raw`;

  if (files.length === 1) {
    url += removePrefix(files[0]) + "?";
  } else {
    let arg = "";

    for (let file of files) {
      arg += removePrefix(file) + ",";
    }

    arg = arg.substring(0, arg.length - 1);
    arg = encodeURIComponent(arg);
    url += `/?files=${arg}&`;
  }

  if (format) {
    url += `algo=${format}&`;
  }

  if (store.state.jwt) {
    url += `auth=${store.state.jwt}&`;
  }

  window.open(url);
}

export async function post(url, content = "", overwrite = false, onupload) {
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

async function postResources(url, content = "", overwrite = false, onupload) {
  url = removePrefix(url);

  let bufferContent;
  if (
    content instanceof Blob &&
    !["http:", "https:"].includes(window.location.protocol)
  ) {
    bufferContent = await new Response(content).arrayBuffer();
  }

  return new Promise((resolve, reject) => {
    let request = new XMLHttpRequest();
    request.open(
      "POST",
      `${baseURL}/api/resources${url}?override=${overwrite}`,
      true
    );
    request.setRequestHeader("X-Auth", store.state.jwt);

    if (typeof onupload === "function") {
      request.upload.onprogress = onupload;
    }

    request.onload = () => {
      if (request.status === 200) {
        resolve(request.responseText);
      } else if (request.status === 409) {
        reject(request.status);
      } else {
        reject(request.responseText);
      }
    };

    request.onerror = () => {
      reject(new Error("001 Connection aborted"));
    };

    request.send(bufferContent || content);
  });
}

function moveCopy(items, copy = false, overwrite = false, rename = false) {
  let promises = [];

  for (let item of items) {
    const from = item.from;
    const to = encodeURIComponent(removePrefix(item.to));
    const url = `${from}?action=${
      copy ? "copy" : "rename"
    }&destination=${to}&override=${overwrite}&rename=${rename}`;
    promises.push(resourceAction(url, "PATCH"));
  }

  return Promise.all(promises);
}

export function move(items, overwrite = false, rename = false) {
  return moveCopy(items, false, overwrite, rename);
}

export function copy(items, overwrite = false, rename = false) {
  return moveCopy(items, true, overwrite, rename);
}

export async function checksum(url, algo) {
  const data = await resourceAction(`${url}?checksum=${algo}`, "GET");
  return (await data.json()).checksums[algo];
}

export function getDownloadURL(file, inline) {
  const params = {
    ...(inline && { inline: "true" }),
  };

  return createURL("api/raw" + file.path, params);
}

export function getPreviewURL(file, size) {
  const params = {
    inline: "true",
    key: Date.parse(file.modified),
  };

  return createURL("api/preview/" + size + file.path, params);
}

export function getSubtitlesURL(file) {
  const params = {
    inline: "true",
  };

  const subtitles = [];
  for (const sub of file.subtitles) {
    subtitles.push(createURL("api/raw" + sub, params));
  }

  return subtitles;
}

export async function usage(url) {
  url = removePrefix(url);

  const res = await fetchURL(`/api/usage${url}`, {});

  return await res.json();
}
