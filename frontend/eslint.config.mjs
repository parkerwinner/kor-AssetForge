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
  // jest.mock() factories are hoisted above imports, so they must use require() (#439).
  {
    files: ["__tests__/**/*.{ts,tsx}"],
    rules: {
      "@typescript-eslint/no-require-imports": "off",
    },
  },
  // React Compiler rules stay errors everywhere except in these files, which
  // predate them. Fix a file, then remove it from this list. See docs/LINTING.md (#439).
  {
    files: [
      "src/app/admin/components/ComplianceReports.tsx",
      "src/app/admin/components/KYCReview.tsx",
      "src/app/admin/components/UserManagement.tsx",
      "src/app/analytics/page.tsx",
      "src/app/bundles/page.tsx",
      "src/app/escrow/page.tsx",
      "src/app/governance/page.tsx",
      "src/app/transactions/page.tsx",
      "src/app/verification/page.tsx",
      "src/components/FileUpload.tsx",
      "src/components/FilterPanel.tsx",
      "src/components/TransactionFilter.tsx",
      "src/components/charts/AssetAllocationChart.tsx",
      "src/components/escrow/EscrowList.tsx",
      "src/components/governance/ProposalList.tsx",
      "src/components/governance/VotePanel.tsx",
      "src/components/onboarding/OnboardingProvider.tsx",
    ],
    rules: {
      "react-hooks/immutability": "warn",
      "react-hooks/preserve-manual-memoization": "warn",
      "react-hooks/purity": "warn",
      "react-hooks/set-state-in-effect": "warn",
    },
  },
  // Override default ignores of eslint-config-next.
  globalIgnores([
    // Default ignores of eslint-config-next:
    ".next/**",
    "out/**",
    "build/**",
    "next-env.d.ts",
  ]),
]);

export default eslintConfig;
