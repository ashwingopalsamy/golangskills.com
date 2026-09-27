import { getSiteData } from "../../lib/catalog.js";

const siteOrigin = "https://golangskills.com";
export function getStaticPaths() {
  return getSiteData().collections.map((collection) => ({ params: { slug: collection.slug }, props: { collection } }));
}

export function GET({ props }) {
  const { collection } = props;
  const lines = [
    `# ${collection.display_name}`,
    "",
    collection.description,
    "",
    `Canonical collection guide: ${collection.sourceUrl}`,
    "",
    "Install all collections from the canonical repository:",
    "",
    "```sh",
    "npx skills add ashwingopalsamy/golangskills.com",
    "```",
    "",
    "## Skills",
    "",
  ];
  for (const skill of collection.skills) lines.push(`- [${skill.display_name}](${siteOrigin}${skill.route}) — ${skill.description}`);
  lines.push("");
  return new Response(lines.join("\n"), { headers: { "Content-Type": "text/markdown; charset=utf-8" } });
}
