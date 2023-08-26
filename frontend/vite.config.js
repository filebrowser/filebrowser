import { fileURLToPath, URL } from "node:url";

import { defineConfig } from "vite";
import legacy from "@vitejs/plugin-legacy";
import vue2 from "@vitejs/plugin-vue2";
import { compression } from "vite-plugin-compression2";

const plugins = [
  vue2(),
  legacy({
    targets: ["ie >= 11"],
    additionalLegacyPolyfills: ["regenerator-runtime/runtime"],
  }),
  compression({ include: /\.js$/i, deleteOriginalAssets: true }),
];

const resolve = {
  alias: {
    vue: "vue/dist/vue.esm.js",
    "@/": fileURLToPath(new URL("./src/", import.meta.url)),
  },
};

// https://vitejs.dev/config/
export default defineConfig(({ command }) => {
  if (command === "serve") {
    return {
      plugins,
      resolve,
      server: {
        proxy: {
          "/api": "http://127.0.0.1:8080",
        },
      },
    };
  } else {
    // command === 'build'
    return {
      plugins,
      resolve,
      base: "",
      build: {
        rollupOptions: {
          input: {
            index: fileURLToPath(
              new URL(`./public/index.html`, import.meta.url)
            ),
          },
        },
      },
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
    };
  }
});
