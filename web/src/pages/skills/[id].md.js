import { readFile } from "node:fs/promises";
import { resolve } from "node:path";
import { getSiteData, repositoryRoot } from "../../lib/catalog.js";

export function getStaticPaths() {
  return getSiteData().skills.map((skill) => ({ params: { id: skill.name }, props: { skill } }));
}

export async function GET({ props }) {
  const content = await readFile(resolve(repositoryRoot, props.skill.path, "SKILL.md"), "utf8");
  return new Response(content, { headers: { "Content-Type": "text/markdown; charset=utf-8" } });
}
