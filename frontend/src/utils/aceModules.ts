// Bundle Ace's themes / modes / workers locally and register them with
// ace.config.setModuleUrl, so the editor never needs to reach a CDN.
//
// The upstream code pointed Ace's basePath at jsdelivr — fine on a dev
// laptop, useless on a shop-LAN Pi that may be offline. With the modules
// bundled, theme switching works without internet and there is no FOUC
// while the CDN file streams in.
//
// Vite turns each ?url import into a hashed asset URL, so the cost is one
// extra <script> per module the user actually selects (Ace fetches them
// on demand via setTheme/setMode). Nothing is downloaded eagerly.

import ace from "ace-builds";

type UrlMap = Record<string, string>;

const themeUrls = import.meta.glob(
  "../../node_modules/ace-builds/src-min-noconflict/theme-*.js",
  { eager: true, query: "?url", import: "default" }
) as UrlMap;

const modeUrls = import.meta.glob(
  "../../node_modules/ace-builds/src-min-noconflict/mode-*.js",
  { eager: true, query: "?url", import: "default" }
) as UrlMap;

const workerUrls = import.meta.glob(
  "../../node_modules/ace-builds/src-min-noconflict/worker-*.js",
  { eager: true, query: "?url", import: "default" }
) as UrlMap;

// Ace extensions — most notably ext-searchbox, which Ace dynamically
// loads when the user opens Find (Ctrl-F). Without this glob, Ace
// requests "/files/ext-searchbox.js" relative to the page, the SPA
// catch-all returns index.html, and the editor blows up with
// "Unexpected token '<'" / "Cannot read properties of undefined
// (reading 'Search')". Bundling these the same way as themes/modes
// keeps Find / Goto-line / etc. working on the air-gapped Pi.
const extUrls = import.meta.glob(
  "../../node_modules/ace-builds/src-min-noconflict/ext-*.js",
  { eager: true, query: "?url", import: "default" }
) as UrlMap;

const nameOf = (path: string, prefix: string): string | null => {
  const m = new RegExp(`/${prefix}-([^/]+?)\\.js$`).exec(path);
  return m ? m[1] : null;
};

for (const [path, url] of Object.entries(themeUrls)) {
  const name = nameOf(path, "theme");
  if (name) ace.config.setModuleUrl(`ace/theme/${name}`, url);
}

for (const [path, url] of Object.entries(modeUrls)) {
  const name = nameOf(path, "mode");
  if (name) ace.config.setModuleUrl(`ace/mode/${name}`, url);
}

// Workers register under ace/mode/<lang>_worker — file name is worker-<lang>.js.
for (const [path, url] of Object.entries(workerUrls)) {
  const name = nameOf(path, "worker");
  if (name) ace.config.setModuleUrl(`ace/mode/${name}_worker`, url);
}

// Extensions register under ace/ext/<name> — file name is ext-<name>.js.
for (const [path, url] of Object.entries(extUrls)) {
  const name = nameOf(path, "ext");
  if (name) ace.config.setModuleUrl(`ace/ext/${name}`, url);
}
