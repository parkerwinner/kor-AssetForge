# Linting

ESLint setup for the Next.js 16 / React 19 frontend (issue #439).

## Run it

```bash
cd frontend
npm run lint          # eslint (flat config, lints the whole frontend)
npx eslint --fix .    # apply auto-fixes
```

CI runs `npm run lint` in the **Build Frontend** job, before the type check.
Errors fail the build; warnings don't.

Next.js 16 removed `next lint`, so the script calls the ESLint CLI directly.

## Configuration

There's one config file, `frontend/eslint.config.mjs` (ESLint 9 flat config).
The frontend has no `.eslintrc*`, `.eslintignore` or `eslintConfig` in
`package.json`, so nothing conflicts with it. Don't add legacy config files:
ESLint 9 ignores them.

| Block | Source | Purpose |
| --- | --- | --- |
| `eslint-config-next/core-web-vitals` | `eslint-config-next` 16.2.4 (matches `next` 16.2.4) | Next.js rules (`@next/next/*`) with Core Web Vitals rules as errors, plus the recommended `react`, `react-hooks` and `jsx-a11y` rules. |
| `eslint-config-next/typescript` | same | `typescript-eslint` recommended rules. |
| `src/actions/**` | `eslint-rules/require-use-server.js` | Server Action modules must start with `'use server'` (#423). |
| `__tests__/**` | local | Allows `require()`: `jest.mock()` factories are hoisted above imports, so they can't use `import`. |
| Compiler-rule baseline | local | See below. |
| `globalIgnores` | local | `.next/**`, `out/**`, `build/**`, `next-env.d.ts`. |

Keep `eslint-config-next` on the same version as `next` when upgrading.

## Next.js 16 rules

Highlights from `@next/eslint-plugin-next`:

- `@next/next/no-html-link-for-pages`: use `<Link>`, not `<a>`, for internal pages.
- `@next/next/no-img-element`: use `next/image`.
- `@next/next/no-async-client-component`: Client Components can't be `async`.
- `@next/next/no-head-element`, `no-css-tags`, `no-sync-scripts`, `google-font-*`.

## React 19 rules

`eslint-config-next` 16 ships `eslint-plugin-react-hooks` v7, whose recommended
set includes the React Compiler rules. All are errors unless noted:

| Rule | Catches |
| --- | --- |
| `rules-of-hooks` | Hooks called conditionally or outside components. |
| `exhaustive-deps` (warn) | Missing effect/memo dependencies. |
| `purity` | Impure calls during render, such as `Date.now()` or `Math.random()`. |
| `set-state-in-effect` | Synchronous `setState` in an effect body (cascading renders). |
| `set-state-in-render` | `setState` called during render. |
| `immutability` | Mutating props, state, or variables after render. |
| `refs` | Reading or writing `ref.current` during render. |
| `preserve-manual-memoization` | `useMemo`/`useCallback` the compiler can't preserve. |
| `static-components` | Components created inside other components. |
| `use-memo` | Common `useMemo` mistakes. |
| `error-boundaries` | `try`/`catch` around JSX instead of an error boundary. |
| `globals` | Assigning to globals during render. |
| `config`, `gating` | React Compiler configuration problems. |
| `incompatible-library` (warn) | Libraries that are incompatible with memoization. |
| `unsupported-syntax` (warn) | Syntax React Compiler won't support. |

`npx eslint --print-config src/app/page.tsx` shows the full resolved rule set.

## Compiler-rule baseline

React Compiler runs in annotation mode (`'use memo'` opt-in; see
[NEXTJS_16_MIGRATION.md](NEXTJS_16_MIGRATION.md)) because some components that
predate these rules still break them. To keep lint passing in CI without
weakening the rules for new code:

- `react-hooks/purity`, `set-state-in-effect`, `immutability` and
  `preserve-manual-memoization` are **errors everywhere**,
- **except** in the 17 files listed in the baseline block of `eslint.config.mjs`,
  where they're warnings.

New files are never added to the list. To pay down the baseline:

1. Fix the warnings in one listed file.
2. Remove it from the list. Its rules go back to errors.
3. Optionally add `'use memo'` to opt the component into the compiler.

When the list is empty, delete the block and consider switching to
`reactCompiler: true`.

## Disabling a rule

Prefer fixing the code. If a line really needs an exception, disable one rule
on one line and give the reason:

```tsx
// eslint-disable-next-line @next/next/no-img-element -- remote SVG, next/image can't optimise it
<img src={logoUrl} alt="" />
```

ESLint warns about `eslint-disable` comments that no longer suppress anything;
remove them when you see that warning.
