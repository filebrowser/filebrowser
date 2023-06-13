import { fetchURL, removePrefix } from "./utils";
import url from "../utils/url";

export default async function search(base, query) {
  base = removePrefix(base);
  query = encodeURIComponent(query);

  if (!base.endsWith("/")) {
    base += "/";
  }

  let res = await fetchURL(`/api/search${base}?query=${query}`, {});

  let data = await res.json();

  data = data.map((item) => {
    item.url = `/files${base}` + url.encodePath(item.path);

    if (item.dir) {
      item.url += "/";
    }

    return item;
  });

  return data;
}
