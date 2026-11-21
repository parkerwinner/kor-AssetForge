"use client";

import type { ReactNode } from "react";
import { QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { getQueryClient } from "@/lib/query-client";

export function Providers({ children }: { children: ReactNode }) {
  // Not useState: React discards state if the first render suspends with no boundary above.
  const queryClient = getQueryClient();
  return (
    <QueryClientProvider client={queryClient}>
      {children}
      {/* Renders nothing unless NODE_ENV is "development". */}
      <ReactQueryDevtools />
    </QueryClientProvider>
  );
}
