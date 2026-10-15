# HTTP security headers

The Go API applies these response headers to every route through `middleware.SecurityHeaders`:

| Header | Value | Purpose |
| --- | --- | --- |
| `Content-Security-Policy` | `default-src 'self'` with explicit source restrictions | Limits executable content, embedded frames, objects, images, fonts, and API connections to approved sources. |
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains` | Tells browsers to use HTTPS for the service and its subdomains after the first secure response. |
| `X-Content-Type-Options` | `nosniff` | Prevents MIME type sniffing. |
| `X-Frame-Options` | `DENY` | Prevents the API responses from being framed. |
| `X-XSS-Protection` | `1; mode=block` | Enables legacy browser XSS filtering where supported. |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Limits referrer details sent to another origin. |
| `Permissions-Policy` | `camera=(), geolocation=(), microphone=()` | Disables browser access to unused device capabilities. |

The HSTS header assumes production traffic is served over HTTPS. Local development can continue to use HTTP; browsers only enforce HSTS after receiving the header over a secure connection.
