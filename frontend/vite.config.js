import { fileURLToPath, URL } from "node:url";

import { defineConfig } from "vite";
import legacy from "@vitejs/plugin-legacy";
import vue2 from "@vitejs/plugin-vue2";
import { compression } from "vite-plugin-compression2";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue2(),
    legacy({
      targets: ["ie >= 11"],
      additionalLegacyPolyfills: ["regenerator-runtime/runtime"],
    }),
    compression({ include: /\.js$/i, deleteOriginalAssets: true }),
  ],
  resolve: {
    alias: {
      "@/": fileURLToPath(new URL("./src/", import.meta.url)),
    },
  },
  base: "",
  experimental: {
    renderBuiltUrl(filename, { hostType }) {
      if (hostType === "js") {
        return { runtime: `window.__appendStaticUrl("${filename}")` };
        // } else if (hostType === "css") {
        //   return `'[{[ .StaticURL ]}]/${filename}'`;
      } else if (hostType === "html") {
        return `[{[ .StaticURL ]}]/${filename}`;
        // return {
        //   runtime: `window.__appendStaticUrl(${JSON.stringify(filename)})`,
        // };
        // if (hostType === "js") {
        // return { runtime: `window.__toCdnUrl(${JSON.stringify(filename)})` };
      } else {
        return { relative: true };
      }
    },
  },
});
