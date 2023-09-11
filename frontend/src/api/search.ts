import { fetchURL, removePrefix } from "./utils";
import url from "../utils/url";
import type { Item } from "@/types";

export default async function search(base: string, query: string) {
  base = removePrefix(base);
  query = encodeURIComponent(query);

  if (!base.endsWith("/")) {
    base += "/";
  }

  const res = await fetchURL(`/api/search${base}?query=${query}`, {});

  let data = await res.json();

  data = data.map((item: Item) => {
    item.url = `/files${base}` + url.encodePath(item.path);

    if (item.dir) {
      item.url += "/";
    }

    return item;
  });

  return data;
}
