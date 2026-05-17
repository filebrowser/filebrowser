import path from "node:path";
import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import VueI18nPlugin from "@intlify/unplugin-vue-i18n/vite";
import legacy from "@vitejs/plugin-legacy";
import { compression } from "vite-plugin-compression2";

// Legacy bundle (Babel + ES5 polyfills) doubles build time on a Pi —
// PLUGIN_TIMINGS showed 42% of build cost in vite:legacy-post-process
// + 33% in vite:terser. We only need it if operators have IE/very old
// browsers, which is not the case for an internal CNC dashboard. Opt
// in via INCLUDE_LEGACY=1 if a deployment ever needs it.
const includeLegacy = process.env.INCLUDE_LEGACY === "1";

const plugins = [
  vue(),
  VueI18nPlugin({
    include: [path.resolve(__dirname, "./src/i18n/**/*.json")],
  }),
  ...(includeLegacy
    ? [
        legacy({
          // defaults already drop IE support
          targets: ["defaults"],
        }),
      ]
    : []),
  compression({ include: /\.js$/, deleteOriginalAssets: false }),
];

const resolve = {
  alias: {
    // vue: "@vue/compat",
    "@/": `${path.resolve(__dirname, "src")}/`,
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
          "/api/command": {
            target: "ws://127.0.0.1:8080",
            ws: true,
          },
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
        // esbuild is ~10x faster than terser at comparable compression
        // for our bundle. Saves 30-60s on a Pi 4 build; meaningful for
        // operators who rebuild on the device. terser still available
        // via INCLUDE_LEGACY=1 (plugin-legacy pulls it in regardless).
        minify: "esbuild",
        // chunkSizeWarningLimit raised: a couple of our chunks
        // (codemirror + three) genuinely belong as single units; the
        // 500 KB default is noise on this app.
        chunkSizeWarningLimit: 1500,
        rollupOptions: {
          input: {
            index: path.resolve(__dirname, "./public/index.html"),
          },
          output: {
            manualChunks: (id) => {
              // bundle dayjs files in a single chunk
              // this avoids having small files for each locale
              if (id.includes("dayjs/")) {
                return "dayjs";
                // bundle i18n in a separate chunk
              } else if (id.includes("i18n/")) {
                return "i18n";
              }
            },
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
