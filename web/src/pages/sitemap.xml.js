import { getSiteData } from "../lib/catalog.js";
import { xmlSitemap } from "../lib/discovery.js";

export function GET() {
  return new Response(xmlSitemap(getSiteData()), {
    headers: { "Content-Type": "application/xml; charset=utf-8" },
  });
}
