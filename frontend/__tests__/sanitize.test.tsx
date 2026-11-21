import { render } from "@testing-library/react";
import { sanitizeUserContent } from "@/lib/sanitize";
import { UserContent } from "@/components/UserContent";

describe("sanitizeUserContent", () => {
  it("keeps the formatting the backend policy allows", () => {
    const html =
      "<p><strong>Bold</strong> <em>it</em><br></p><ul><li>a</li></ul><ol><li>b</li></ol>" +
      "<blockquote>q</blockquote><pre><code>x</code></pre>";
    expect(sanitizeUserContent(html)).toBe(html);
  });

  it.each([
    ["<p>hi</p><script>alert(1)</script>", "<p>hi</p>"],
    ['<img src="x" onerror="alert(1)">', ""],
    ['<p onclick="alert(1)" style="color:red" class="x">t</p>', "<p>t</p>"],
    ['<svg><script>alert(1)</script></svg>', ""],
    ['<iframe src="https://evil.example"></iframe>', ""],
    ["<div><span>text kept</span></div>", "text kept"],
  ])("strips unsafe markup from %s", (dirty, clean) => {
    expect(sanitizeUserContent(dirty)).toBe(clean);
  });

  it.each([
    "javascript:alert(1)",
    " javascript:alert(1)",
    "JaVaScRiPt:alert(1)",
    "data:text/html;base64,PHNjcmlwdD5hbGVydCgxKTwvc2NyaXB0Pg==",
    "/relative/path",
  ])("drops href %s", (href) => {
    expect(sanitizeUserContent(`<a href="${href}">x</a>`)).toBe("<a>x</a>");
  });

  it.each(["https://example.com", "http://example.com", "mailto:a@example.com"])(
    "keeps href %s and adds rel=nofollow",
    (href) => {
      expect(sanitizeUserContent(`<a href="${href}" target="_blank" rel="opener">x</a>`)).toBe(
        `<a href="${href}" rel="nofollow">x</a>`,
      );
    },
  );

  it("returns an empty string when no DOM is available", async () => {
    await jest.isolateModulesAsync(async () => {
      jest.doMock("dompurify", () => ({ __esModule: true, default: { isSupported: false } }));
      const { sanitizeUserContent: serverSanitize } = await import("@/lib/sanitize");
      expect(serverSanitize("<p>hi</p><script>alert(1)</script>")).toBe("");
    });
  });
});

describe("UserContent", () => {
  it("renders sanitized HTML", () => {
    const { container } = render(
      <UserContent html='<p>Hello &amp; welcome</p><img src="x" onerror="alert(1)">' className="text-sm" />,
    );
    const root = container.firstElementChild as HTMLElement;
    expect(root).toHaveClass("text-sm");
    expect(root.innerHTML).toBe("<p>Hello &amp; welcome</p>");
    expect(root).toHaveTextContent("Hello & welcome");
    expect(container.querySelector("img")).toBeNull();
  });
});
