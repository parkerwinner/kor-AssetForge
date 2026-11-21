import DOMPurify from "dompurify";

// Mirrors the backend bluemonday policy (backend/utils/sanitizer.go), so content
// that bypasses the API, such as IPFS metadata, gets the same allowlist.
const USER_CONTENT_CONFIG = {
  ALLOWED_TAGS: ["p", "br", "strong", "em", "ul", "ol", "li", "blockquote", "code", "pre", "a"],
  ALLOWED_ATTR: ["href"],
  ALLOWED_URI_REGEXP: /^(?:https?|mailto):/i,
};

if (DOMPurify.isSupported) {
  DOMPurify.addHook("afterSanitizeAttributes", (node) => {
    if (node.tagName === "A" && node.hasAttribute("href")) {
      node.setAttribute("rel", "nofollow");
    }
  });
}

/**
 * Sanitizes user-generated HTML for rendering. Returns an empty string where
 * there is no DOM (server render), so unsanitized markup is never emitted.
 */
export function sanitizeUserContent(dirty: string): string {
  if (!DOMPurify.isSupported) return "";
  return DOMPurify.sanitize(dirty, USER_CONTENT_CONFIG);
}
