import js from "@eslint/js";

const eslintConfig = [
  js.configs.recommended,
  {
    ignores: ["out/", ".next/", "node_modules/", "scripts/"],
  },
];

export default eslintConfig;
