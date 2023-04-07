import { fetchURL, fetchJSON, removePrefix, createURL } from "./utils";

export async function list() {
  return fetchJSON("/api/shares");
}

export async function get(url) {
  url = removePrefix(url);
  return fetchJSON(`/api/share${url}`);
}

export async function remove(hash) {
  await fetchURL(`/api/share/${hash}`, {
    method: "DELETE",
  });
}

export async function create(
  config = {
    url: "",
    password: "",
    expires: "",
    unit: "hours",
    custom: false,
    customLink: "",
  }
) {
  let {
    url,
    password = "",
    expires = "",
    unit = "hours",
    custom = false,
    customLink = "",
  } = config;
  url = removePrefix(url);
  url = `/api/share${url}`;
  if (expires !== "") {
    url += `?expires=${expires}&unit=${unit}`;
  }
  let body = {
    custom,
    customLink: customLink,
  };
  if (password !== "" || expires !== "" || unit !== "hours") {
    Object.assign(body, { password: password, expires: expires, unit: unit });
  }
  body = JSON.stringify(body);
  return fetchJSON(url, {
    method: "POST",
    body: body,
  });
}

export function getShareURL(share) {
  return createURL("share/" + share.hash, {}, false);
}
