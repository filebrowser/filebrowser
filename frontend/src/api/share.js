import { fetchURL, fetchJSON, removePrefix, createURL } from "./utils";

export async function list() {
  return fetchJSON("/api/shares");
}

export async function get(url) {
  url = removePrefix(url);
  return fetchJSON(`/api/share${url}`);
}

export async function remove(hash) {
  const res = await fetchURL(`/api/share/${hash}`, {
    method: "DELETE",
  });

  if (res.status !== 200) {
    throw new Error(res.status);
  }
}

export async function create(url, password = "", expires = "", unit = "hours") {
  url = removePrefix(url);
  url = `/api/share${url}`;
  if (expires !== "") {
    url += `?expires=${expires}&unit=${unit}`;
  }
  let body = "{}";
  if (password != "" || expires !== "" || unit !== "hours") {
    body = JSON.stringify({ password: password, expires: expires, unit: unit });
  }
  return fetchJSON(url, {
    method: "POST",
    body: body,
  });
}

export function getShareURL(share) {
  return createURL("share/" + share.hash, {}, false);
}
