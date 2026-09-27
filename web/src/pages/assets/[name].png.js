import { readFile } from "node:fs/promises";
import { resolve } from "node:path";
import { repositoryRoot } from "../../lib/catalog.js";

export function getStaticPaths() {
  return ["engineering", "distributed-systems", "fintech"].map((name) => ({ params: { name: `${name}-logo` } }));
}

export async function GET({ params }) {
  const image = await readFile(resolve(repositoryRoot, "internal/corpus/assets/generated", `${params.name}.png`));
  return new Response(image, { headers: { "Content-Type": "image/png" } });
}
