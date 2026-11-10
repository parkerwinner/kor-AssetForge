import { defineConfig, globalIgnores } from "eslint/config";
import nextVitals from "eslint-config-next/core-web-vitals";
import nextTs from "eslint-config-next/typescript";
import requireUseServer from "./eslint-rules/require-use-server.js";

const eslintConfig = defineConfig([
  ...nextVitals,
  ...nextTs,
  // Server Actions live in src/actions and must be 'use server' modules (#423).
  {
    files: ["src/actions/**/*.{ts,tsx,js,jsx}"],
    plugins: {
      local: { rules: { "require-use-server": requireUseServer } },
    },
    rules: {
      "local/require-use-server": "error",
    },
  },
  // Rules for Next.js 16 and React 19 compatibility (#439)
  {
    rules: {
      "react/no-unescaped-entities": "off",
      "react-hooks/purity": "warn",
      "react-hooks/set-state-in-effect": "warn",
      "react-hooks/immutability": "warn",
      "react-hooks/preserve-manual-memoization": "warn",
      "@typescript-eslint/no-unused-vars": [
        "warn",
        {
          argsIgnorePattern: "^_",
          varsIgnorePattern: "^_",
          caughtErrorsIgnorePattern: "^_",
        },
      ],
      "@typescript-eslint/no-explicit-any": "warn",
    },
  },
  // Test files overrides
  {
    files: ["__tests__/**/*.{ts,tsx,js,jsx}", "**/*.test.{ts,tsx,js,jsx}", "**/*.spec.{ts,tsx,js,jsx}"],
    rules: {
      "@typescript-eslint/no-require-imports": "off",
    },
  },
  // Override default ignores of eslint-config-next.
  globalIgnores([
    // Default ignores of eslint-config-next:
    ".next/**",
    "out/**",
    "build/**",
    "coverage/**",
    "next-env.d.ts",
    "node_modules/**",
  ]),
]);

export default eslintConfig;
