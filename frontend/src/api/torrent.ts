import { fetchURL, fetchJSON, removePrefix, createURL } from "./utils";

export async function fetchDefaultOptions() {
  const url = `/api/torrent`
  return fetchJSON(url);
}

export async function makeTorrent(
  url: string,
  announces: string[],
  comment: string,
  date: boolean,
  name: string,
  pieceLen: number,
  privateFlag: boolean,
  r2Flag: boolean,
  source: string,
  webSeeds: string[]
) {
  // Construct the URL for torrent creation API
  url = removePrefix(url);
  url = `/api/torrent${url}`;

  let body = "{}";
  if (announces.length > 0) {
    body = JSON.stringify({
      announces: announces,
      comment: comment,
      date: date,
      name: name,
      pieceLen: pieceLen,
      private: privateFlag,
      r2: r2Flag,
      source: source,
      webSeeds: webSeeds,
    });
  }

  // Send POST request to create the torrent
  return fetchJSON(url, {
    method: "POST",
    body: body,
  });
}

export function publish(
  url: string
) {
  url = removePrefix(url);
  url = `/api/publish${url}`;

  let body = "{}";

  return fetchJSON(url, {
    method: "POST",
    body: body,
  });
}