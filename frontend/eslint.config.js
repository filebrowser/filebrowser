import pluginVue from "eslint-plugin-vue";
import {
  defineConfigWithVueTs,
  vueTsConfigs,
} from "@vue/eslint-config-typescript";
import prettierConfig from "@vue/eslint-config-prettier";

export default defineConfigWithVueTs(
  {
    name: "app/files-to-lint",
    files: ["**/*.{ts,mts,tsx,vue}"],
  },
  {
    name: "app/files-to-ignore",
    ignores: ["**/dist/**", "**/dist-ssr/**", "**/coverage/**"],
  },
  pluginVue.configs["flat/essential"],
  vueTsConfigs.recommended,
  prettierConfig,
  {
    rules: {
      // Note: you must disable the base rule as it can report incorrect errors
      "@typescript-eslint/no-unused-expressions": "off",
      // TODO: theres too many of these from before ts
      "@typescript-eslint/no-explicit-any": "off",
      // TODO: finish the ts conversion
      "vue/block-lang": "off",
      "vue/multi-word-component-names": "off",
      "vue/no-mutating-props": [
        "error",
        {
          shallowOnly: true,
        },
      ],
    },
  }
);
