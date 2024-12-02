import pluginVue from "eslint-plugin-vue";
import vueTsEslintConfig from "@vue/eslint-config-typescript";
import prettierConfig from "@vue/eslint-config-prettier";

export default [
  {
    name: "app/files-to-lint",
    files: ["**/*.{ts,mts,tsx,vue}"],
  },

  {
    name: "app/files-to-ignore",
    ignores: ["**/dist/**", "**/dist-ssr/**", "**/coverage/**"],
  },

  ...pluginVue.configs["flat/essential"],
  ...vueTsEslintConfig(),
  prettierConfig,

  {
    rules: {
      "no-var": "warn",
      "prefer-const": "warn",
      // Note: you must disable the base rule as it can report incorrect errors
      "no-unused-expressions": "off",
      "@typescript-eslint/no-unused-expressions": "warn",
      "@typescript-eslint/no-explicit-any": "warn",
      "@typescript-eslint/ban-ts-comment": "warn",
      "vue/block-lang": "warn",
      "vue/multi-word-component-names": "off",
      "vue/no-mutating-props": [
        "error",
        {
          shallowOnly: true,
        },
      ],
    },
  },
];
