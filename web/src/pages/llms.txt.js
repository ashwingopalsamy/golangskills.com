import { readFile } from "node:fs/promises";
import { resolve } from "node:path";
import { repositoryRoot } from "../lib/catalog.js";

export async function GET() {
  const content = await readFile(resolve(repositoryRoot, "llms.txt"), "utf8");
  return new Response(content, { headers: { "Content-Type": "text/plain; charset=utf-8" } });
}
