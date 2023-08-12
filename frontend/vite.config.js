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
      vue: "vue/dist/vue.esm.js",
      "@/": fileURLToPath(new URL("./src/", import.meta.url)),
    },
  },
  base: "",
  experimental: {
    renderBuiltUrl(filename, { hostType }) {
      if (hostType === "js") {
        return { runtime: `window.__prependStaticUrl("${filename}")` };
      } else if (hostType === "html") {
        return `[{[ .StaticURL ]}]/${filename}`;
      } else {
        return { relative: true };
      }
    },
  },
});
