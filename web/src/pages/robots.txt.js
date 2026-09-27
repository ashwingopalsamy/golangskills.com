export function GET() {
  return new Response("User-agent: *\nAllow: /\n\nSitemap: https://golangskills.com/sitemap.xml\n", {
    headers: { "Content-Type": "text/plain; charset=utf-8" },
  });
}
