"use client";

import { useSyncExternalStore } from "react";
import { sanitizeUserContent } from "@/lib/sanitize";

const subscribe = () => () => {};

interface UserContentProps {
  html: string;
  className?: string;
}

/** Renders user-generated HTML after sanitizing it. Use for any user-authored text. */
export function UserContent({ html, className }: UserContentProps) {
  // DOMPurify needs a DOM: the server render and hydration pass emit an empty
  // element, then the browser render fills in the sanitized HTML.
  const isBrowser = useSyncExternalStore(subscribe, () => true, () => false);
  return (
    <div
      className={className}
      dangerouslySetInnerHTML={{ __html: isBrowser ? sanitizeUserContent(html) : "" }}
    />
  );
}
