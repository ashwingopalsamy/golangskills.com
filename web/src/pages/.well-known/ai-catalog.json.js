import { agentCatalog } from "../../lib/discovery.js";
import { getSiteData } from "../../lib/catalog.js";

export function GET() {
  return new Response(`${JSON.stringify(agentCatalog(getSiteData()), null, 2)}\n`, {
    headers: { "Content-Type": "application/json; charset=utf-8" },
  });
}
