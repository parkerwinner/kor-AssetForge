# React Query Setup

How `@tanstack/react-query` is wired into the Next.js 16 / React 19 frontend
(issue #438).

## Version and React 19 compatibility

| Package | Version | Peer dependencies |
| --- | --- | --- |
| `@tanstack/react-query` | 5.104.3 | `react: ^18 \|\| ^19` |
| `@tanstack/react-query-devtools` | 5.104.3 | `@tanstack/react-query: ^5.104.3`, `react: ^18 \|\| ^19`, `@types/react: ^18 \|\| ^19` |

The installed React is 19.2.4, so both peer ranges are satisfied and npm reports
no peer dependency warnings. No version change is needed. Keep the two packages
on the same version when you upgrade either one; the devtools peer range requires it.

The package was installed but never wired up: there was no `QueryClient` or
`QueryClientProvider`, so any `useQuery` call would throw
`No QueryClient set, use QueryClientProvider to set one`.

## Files

| File | Role |
| --- | --- |
| `frontend/src/lib/query-client.ts` | `getQueryClient()`: the shared `QueryClient` and its defaults. |
| `frontend/src/app/providers.tsx` | `Providers`: a client component with `QueryClientProvider` and `ReactQueryDevtools`. |
| `frontend/src/app/layout.tsx` | Wraps the app in `<Providers>` inside `NextIntlClientProvider`. |

This follows TanStack's
[Advanced Server Rendering guide](https://tanstack.com/query/v5/docs/framework/react/guides/advanced-ssr)
for the App Router:

- **On the server**, `getQueryClient()` returns a new client on every call, so one
  user's cached data never reaches another user's request.
- **In the browser**, it returns one client for the lifetime of the page. The
  provider doesn't use `useState`: if the first render suspends and there's no
  Suspense boundary above it, React discards the state and would create a new
  client.
- Server/browser detection uses `environmentManager.isServer()`. The older
  `isServer` export is deprecated in 5.104.
- `staleTime` defaults to 60 seconds. With a value of 0, data rendered on the
  server would be refetched as soon as the page hydrates.

## Using queries

Queries need a Client Component:

```tsx
"use client";

import { useQuery } from "@tanstack/react-query";
import { assetApi } from "@/lib/asset-api";

export function AssetName({ id }: { id: string }) {
  const { data, isPending, error } = useQuery({
    queryKey: ["asset", id],
    queryFn: () => assetApi.getAssetDetail(id),
  });

  if (isPending) return <span>Loading…</span>;
  if (error) return <span>Couldn't load the asset.</span>;
  return <span>{data.name}</span>;
}
```

Patterns:

- **Query keys** are arrays that start with the resource name, then its
  parameters: `["asset", id]`, `["assets", { page, filters }]`. Every value the
  `queryFn` uses belongs in the key.
- **Mutations**: call `useMutation`, then invalidate what changed:
  `queryClient.invalidateQueries({ queryKey: ["asset", id] })`, with
  `useQueryClient()` to get the client.
- **Overrides**: set `staleTime`, `refetchInterval` or `enabled` on the individual
  query rather than changing the defaults in `query-client.ts`.

## Devtools

`<ReactQueryDevtools />` is mounted inside the provider. Its package entry
renders nothing unless `process.env.NODE_ENV === "development"`, so it appears
under `next dev` as a floating TanStack button and is left out
of `next build` output with no extra code.

## Tests

`frontend/__tests__/query-client.test.tsx` checks that the browser reuses one
client, the server creates a fresh client per call, the default `staleTime` is
set, and a `useQuery` hook resolves under `Providers` on React 19.

```bash
cd frontend
npm test -- query-client
```
