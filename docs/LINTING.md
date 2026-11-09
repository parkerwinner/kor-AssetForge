# ESLint & Code Quality Standards

This document describes the ESLint configuration, rule sets, and Next.js 16 / React 19 migration standards for AssetForge.

## Overview

AssetForge uses ESLint 9's Flat Configuration format (`eslint.config.mjs`) integrating:
- `eslint-config-next/core-web-vitals`: Core Next.js performance and Web Vitals linting rules.
- `eslint-config-next/typescript`: Strict TypeScript rules for Next.js App Router and Server Components.
- Custom Server Action rules (`require-use-server`): Enforces that all actions in `src/actions` declare `'use server'`.
- React 19 compiler compatibility rule adjustments.

## Configuration Structure (`frontend/eslint.config.mjs`)

```javascript
import { defineConfig, globalIgnores } from "eslint/config";
import nextVitals from "eslint-config-next/core-web-vitals";
import nextTs from "eslint-config-next/typescript";
import requireUseServer from "./eslint-rules/require-use-server.js";

const eslintConfig = defineConfig([
  ...nextVitals,
  ...nextTs,
  // Server Actions rules
  {
    files: ["src/actions/**/*.{ts,tsx,js,jsx}"],
    plugins: {
      local: { rules: { "require-use-server": requireUseServer } },
    },
    rules: {
      "local/require-use-server": "error",
    },
  },
  // Rules for Next.js 16 and React 19 compatibility
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
  // Test file overrides
  {
    files: ["__tests__/**/*.{ts,tsx,js,jsx}", "**/*.test.{ts,tsx,js,jsx}"],
    rules: {
      "@typescript-eslint/no-require-imports": "off",
    },
  },
  // Global ignore list
  globalIgnores([
    ".next/**",
    "out/**",
    "build/**",
    "coverage/**",
    "next-env.d.ts",
    "node_modules/**",
  ]),
]);

export default eslintConfig;
```

## Running the Linter

```bash
cd frontend
npm run lint
```

To run with automatic fixes:
```bash
npm run lint -- --fix
```

## React 19 Considerations

1. **Ref as Prop**: `forwardRef` is no longer mandatory; components may accept `ref` directly as a prop in React 19.
2. **Compiler Purity**: New compiler checks (`react-hooks/purity`, `react-hooks/set-state-in-effect`) highlight state updates and impure calls during render.
3. **Server Actions**: All mutation endpoints in `src/actions/` must explicitly state `'use server'`.
