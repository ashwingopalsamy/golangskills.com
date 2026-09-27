import { getSiteData } from "../lib/catalog.js";

const siteOrigin = "https://golangskills.com";

export function GET() {
  const { collections, skills } = getSiteData();
  const lines = [
    "# Go Skills for Production Systems",
    "",
    "> Evidence-backed Go skills for AI coding agents, organized across production engineering, distributed systems, and fintech.",
    "",
    "Install from the canonical repository:",
    "",
    "```sh",
    "npx skills add ashwingopalsamy/golangskills.com",
    "```",
    "",
    `Installation guide: ${siteOrigin}/skill.html`,
    "",
    "## Collections",
    "",
  ];
  for (const collection of collections) lines.push(`- [${collection.display_name}](${siteOrigin}${collection.route}) — ${collection.description} (${collection.skill_count} skills)`);
  lines.push("", "## Skills", "");
  for (const skill of skills) lines.push(`- [${skill.display_name}](${siteOrigin}${skill.route}) — ${skill.description}`);
  lines.push("", `Canonical source: https://github.com/ashwingopalsamy/golangskills.com`, "");
  return new Response(lines.join("\n"), { headers: { "Content-Type": "text/markdown; charset=utf-8" } });
}
