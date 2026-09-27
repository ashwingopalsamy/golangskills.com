import { getSiteData } from "../lib/catalog.js";
import { markdownSitemap } from "../lib/discovery.js";

export function GET() {
  return new Response(markdownSitemap(getSiteData()), {
    headers: { "Content-Type": "text/markdown; charset=utf-8" },
  });
}
