import { getSiteData } from "../lib/catalog.js";
import { fullCatalogText } from "../lib/discovery.js";

export function GET() {
  return new Response(fullCatalogText(getSiteData()), { headers: { "Content-Type": "text/plain; charset=utf-8" } });
}
