import { fetchURL, fetchJSON } from "./utils";

export function get() {
  return fetchJSON(`/api/settings`, {});
}

export async function update(settings) {
  await fetchURL(`/api/settings`, {
    method: "PUT",
    body: JSON.stringify(settings),
  });
}
