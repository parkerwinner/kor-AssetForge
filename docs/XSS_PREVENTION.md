# XSS Prevention

How kor-AssetForge keeps user-generated content from running script in the
browser (issue #437). User content is sanitized twice: once by the backend
when it's stored, and again by the frontend when it's rendered.

## Layers

| Layer | Where | What it does |
| --- | --- | --- |
| Backend input sanitization | `backend/utils/sanitizer.go` (bluemonday), applied to every request string by `validator.SanitizeStruct` | Strips everything outside the allowlist below before it reaches the database. |
| React escaping | Every `{value}` in JSX | Renders strings as text, never as markup. This is the default for all content. |
| Frontend HTML sanitization | `frontend/src/lib/sanitize.ts` (DOMPurify), rendered by `frontend/src/components/UserContent.tsx` | Sanitizes user-authored text that's rendered as HTML, including content that never passed the backend (IPFS metadata). |

The frontend doesn't rely on the backend alone. Data can reach the UI without
passing the API's validator, for example NFT metadata fetched from IPFS, rows
written before the policy existed, or a compromised upstream. So the frontend
sanitizes on its own.

## Allowed HTML

The frontend policy mirrors the backend's bluemonday policy:

| | Allowed |
| --- | --- |
| Elements | `p`, `br`, `strong`, `em`, `ul`, `ol`, `li`, `blockquote`, `code`, `pre`, `a` |
| Attributes | `href` on `a` only |
| URL schemes | `http:`, `https:`, `mailto:` (relative, `javascript:` and `data:` URLs are removed) |
| Links | `rel="nofollow"` is added to every link that keeps its `href` |

Everything else is removed: `<script>`, `<style>`, `<iframe>`, `<svg>`, `<img>`,
event handlers (`onclick`, `onerror`, …), `style` and `class` attributes. Text
inside other removed elements, such as `<div>` or `<span>`, is kept as plain text.

If you change the allowlist, change both `backend/utils/sanitizer.go` and
`frontend/src/lib/sanitize.ts`, and update this table.

## Rendering user content in the frontend

Render user-authored text (asset, bundle and proposal descriptions, NFT metadata)
with `UserContent`:

```tsx
import { UserContent } from "@/components/UserContent";

<UserContent html={asset.description} className="text-sm text-muted-foreground" />
```

`UserContent` renders a `<div>`, so don't place it inside a `<p>`.

Rules:

- Don't use `dangerouslySetInnerHTML` anywhere else. `UserContent` is the only
  place that sets HTML, and it always goes through `sanitizeUserContent`.
- Plain values that are never HTML (names, codes, addresses, numbers) stay as
  `{value}` in JSX. React escapes them.
- Validate user input before putting it into `href`, `src` or `style` props.
  React 19 blocks `javascript:` URLs in those props, but it doesn't check other
  schemes or destinations.

### Server rendering

DOMPurify needs a DOM, and in Node it has no `sanitize` function. So:

- `sanitizeUserContent()` returns an empty string when no DOM is available. It
  fails closed and never returns unsanitized markup.
- `UserContent` renders an empty element on the server and during hydration, then
  fills in the sanitized HTML in the browser (`useSyncExternalStore` with a
  `false` server snapshot). Server and client markup match, so there's no
  hydration mismatch.

The components that use it load their data in the browser, so they don't lose
any server-rendered content.

## Tests

`frontend/__tests__/sanitize.test.tsx` covers the allowlist, script and
event-handler removal, `javascript:`/`data:`/relative URL removal, `rel="nofollow"`,
the server-side empty result, and `UserContent` rendering.

```bash
cd frontend
npm test -- sanitize
```

## Dependency

`dompurify` `^3.4.17`. Every published DOMPurify advisory affects versions
up to 3.4.16, so keep the floor at 3.4.17 or later.
