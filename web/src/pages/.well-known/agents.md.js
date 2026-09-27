import { agentGuide } from "../../lib/discovery.js";
import { getSiteData } from "../../lib/catalog.js";

export function GET() {
  return new Response(agentGuide(getSiteData()), { headers: { "Content-Type": "text/markdown; charset=utf-8" } });
}
