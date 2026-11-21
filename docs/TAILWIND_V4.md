# Tailwind CSS v4

How Tailwind CSS v4 is set up in the frontend, and the v3 → v4 changes to
watch for (issue #440). Installed: `tailwindcss` 4.2.4, `@tailwindcss/postcss` 4.2.4.

## Setup

v4 is configured in CSS. There's no `tailwind.config.js` / `.ts`.

| File | Contents |
| --- | --- |
| `frontend/postcss.config.mjs` | The only plugin is `@tailwindcss/postcss`. v4 handles imports and vendor prefixes itself, so `postcss-import` and `autoprefixer` aren't needed. |
| `frontend/src/app/globals.css` | `@import "tailwindcss";` (replaces v3's `@tailwind base/components/utilities`), then `tw-animate-css` and `shadcn/tailwind.css`. |
| | `@custom-variant dark (&:is(.dark *));`: class-based dark mode. |
| | `@theme inline { … }`: design tokens (colors, radii, fonts) mapped to the CSS variables in `:root` / `.dark`. |
| `frontend/components.json` | `"tailwind": { "config": "" }`: shadcn/ui is in CSS-config mode too. |

This matches the setup in the Next.js 16 CSS guide
(`node_modules/next/dist/docs/01-app/01-getting-started/11-css.md`).

### Why there's no `tailwind.config.ts`

v4 doesn't detect JavaScript config files. One only loads if a stylesheet adds
`@config "../tailwind.config.ts";`. Everything a v3 config held lives in CSS:

| v3 `tailwind.config.js` | v4 equivalent |
| --- | --- |
| `theme.extend.colors`, `borderRadius`, `fontFamily` | `@theme` / `@theme inline` variables in `globals.css` |
| `darkMode: "class"` | `@custom-variant dark (&:is(.dark *));` |
| `content: [...]` | Automatic source detection; add `@source "<path>";` for paths it skips |
| `plugins: [...]` | `@plugin "<package>";` |
| `safelist` | `@source inline("<classes>");` (`safelist`, `corePlugins` and `separator` aren't supported in v4) |

Add new tokens to `@theme inline` in `globals.css`, not to a JS config.

## Class generation ("JIT")

v4 has no separate JIT mode: it always generates only the classes it finds. It
scans the project automatically. It skips `.gitignore`d paths (such as
`node_modules`, `.next`), binary files and CSS files. No `content` array is
needed.

Classes must appear as complete strings in source. Build them from maps, not
string concatenation:

```tsx
// Found by the scanner
const tone = { ok: "text-green-600", bad: "text-red-600" }[status];
// Not found: the class never appears whole in source
const tone = `text-${color}-600`;
```

## v3 → v4 changes

### Removed utilities

The upgrade guide lists these as removed. Checked against the installed 4.2.4:

| v3 | v4 | In 4.2.4 |
| --- | --- | --- |
| `bg-opacity-*`, `text-opacity-*`, `border-opacity-*`, `divide-opacity-*`, `ring-opacity-*`, `placeholder-opacity-*` | Opacity modifier: `bg-black/50`, `text-black/50`, … | **Generates nothing.** The style is silently lost. |
| `flex-shrink-*` / `flex-grow-*` | `shrink-*` / `grow-*` | Still compiles, but it's undocumented. Use the v4 name. |
| `overflow-ellipsis` | `text-ellipsis` | Still compiles, but it's undocumented. Use the v4 name. |
| `decoration-slice` / `decoration-clone` | `box-decoration-slice` / `box-decoration-clone` | Still compiles, but it's undocumented. Use the v4 name. |

### Renamed scales

The bare names still work, but the `-sm` steps moved down:

| v3 | v4 |
| --- | --- |
| `shadow-sm` / `shadow` | `shadow-xs` / `shadow-sm` |
| `drop-shadow-sm` / `drop-shadow` | `drop-shadow-xs` / `drop-shadow-sm` |
| `blur-sm` / `blur` | `blur-xs` / `blur-sm` |
| `backdrop-blur-sm` / `backdrop-blur` | `backdrop-blur-xs` / `backdrop-blur-sm` |
| `rounded-sm` / `rounded` | `rounded-xs` / `rounded-sm` |
| `outline-none` | `outline-hidden` (v4's `outline-none` sets `outline-style: none`) |
| `ring` (3px) | `ring-3` (bare `ring` is now 1px) |

This project started on v4 and its shadcn/ui components are written for the v4
scale, so existing `shadow-sm` / `rounded-sm` usages are intentional. Use the v4
names when writing new code.

### Syntax

| v3 | v4 |
| --- | --- |
| `!flex` (important prefix) | `flex!` (the leading `!` still works but is deprecated) |
| `bg-[--brand]` | `bg-(--brand)` |
| `grid-cols-[1fr,auto]` | `grid-cols-[1fr_auto]` (commas inside functions like `min()` are still fine) |
| `@layer utilities { .x {…} }` for custom utilities | `@utility x {…}` |
| `theme(colors.red.500)` in CSS | `var(--color-red-500)` |

### Default style changes

- Default border color is `currentColor` (was gray-200). `globals.css` sets
  `* { @apply border-border }`, so borders keep the theme color.
- Default `ring` width is 1px and its color is `currentColor`.
- Buttons use `cursor: default`.
- `hover:` only applies on devices that support hover.

## Fixes made for #440

The frontend was audited for everything in the tables above:

| File | v3 class | v4 class | Effect |
| --- | --- | --- | --- |
| `src/app/admin/components/UserManagement.tsx` | `bg-black bg-opacity-50` | `bg-black/50` | `bg-opacity-50` generated nothing, so the "Change Role" modal backdrop was solid black instead of 50% black. |
| `src/app/transactions/page.tsx` | `!block` | `block!` | Same output; moves off the deprecated important prefix. |

No other removed utilities, v3 important prefixes, `[--var]` shorthands,
comma-separated arbitrary values, `@tailwind` directives or `theme()` calls
remain in `frontend/src`.

## Checking classes

To check whether Tailwind generates a class, compile it and look for the selector:

```bash
cd frontend
npm run build
grep -lF '.bg-black\/50' .next/static/chunks/*.css    # prints a file: generated
grep -lF '.bg-opacity-50' .next/static/chunks/*.css   # prints nothing: not generated
```

## References

- Tailwind upgrade guide: https://tailwindcss.com/docs/upgrade-guide
- Detecting classes: https://tailwindcss.com/docs/detecting-classes-in-source-files
- Theme variables: https://tailwindcss.com/docs/theme
