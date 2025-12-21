import { fetchURL, removePrefix, StatusError } from "./utils";
import url from "../utils/url";

export default async function search(
  base: string,
  query: string,
  signal: AbortSignal,
  callback: (item: ResourceItem) => void
) {
  base = removePrefix(base);
  query = encodeURIComponent(query);

  if (!base.endsWith("/")) {
    base += "/";
  }

  const res = await fetchURL(`/api/search${base}?query=${query}`, { signal });
  if (!res.body) {
    throw new StatusError("000 No connection", 0);
  }
  try {
    // Try streaming approach first (modern browsers)
    if (res.body && typeof res.body.pipeThrough === "function") {
      const reader = res.body.pipeThrough(new TextDecoderStream()).getReader();
      let buffer = "";
      while (true) {
        const { done, value } = await reader.read();
        if (value) {
          buffer += value;
        }
        const lines = buffer.split(/\n/);
        let lastLine = lines.pop();
        // Save incomplete last line
        if (!lastLine) {
          lastLine = "";
        }
        buffer = lastLine;

        for (const line of lines) {
          if (line) {
            const item = JSON.parse(line) as ResourceItem;
            item.url = `/files${base}` + url.encodePath(item.path);
            if (item.isDir) {
              item.url += "/";
            }
            callback(item);
          }
        }
        if (done) break;
      }
    } else {
      // Fallback for browsers without streaming support (e.g., Safari)
      const text = await res.text();
      const lines = text.split(/\n/);
      for (const line of lines) {
        if (line) {
          const item = JSON.parse(line) as ResourceItem;
          item.url = `/files${base}` + url.encodePath(item.path);
          if (item.isDir) {
            item.url += "/";
          }
          callback(item);
        }
      }
    }
  } catch (e) {
    // Check if the error is an intentional cancellation
    if (e instanceof Error && e.name === "AbortError") {
      throw new StatusError("000 No connection", 0, true);
    }
    throw e;
  }
}
