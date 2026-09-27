import { readFileSync } from "node:fs";
import { resolve } from "node:path";

export const repositoryRoot = resolve(process.cwd(), "..");
const catalog = JSON.parse(readFileSync(resolve(repositoryRoot, "catalog/catalog.json"), "utf8"));

const collectionSlugs = new Map([
  ["engineering-skills-for-go", "engineering"],
  ["distributed-systems-skills-for-go", "distributed-systems"],
  ["fintech-skills-for-go", "fintech"],
]);

const collections = catalog.collections.map((collection) => {
  const slug = collectionSlugs.get(collection.name);
  if (!slug) throw new Error(`No site route configured for collection ${collection.name}`);
  return {
    ...collection,
    slug,
    route: `/collections/${slug}/`,
    logoPath: `/assets/${slug}-logo.png`,
    sourceUrl: `https://github.com/ashwingopalsamy/golangskills.com/blob/main/docs/collections/${slug}.md`,
    skills: [],
  };
});
const collectionByName = new Map(collections.map((collection) => [collection.name, collection]));

const skills = catalog.skills.map((skill) => {
  const collection = collectionByName.get(skill.collection);
  if (!collection) throw new Error(`Skill ${skill.name} references unknown collection ${skill.collection}`);
  const metadataPath = resolve(repositoryRoot, skill.path, "skill.json");
  const instructionsPath = resolve(repositoryRoot, skill.path, "SKILL.md");
  const metadata = JSON.parse(readFileSync(metadataPath, "utf8"));
  const instructions = readFileSync(instructionsPath, "utf8");
  const body = instructions.replace(/^---\r?\n[\s\S]*?\r?\n---\r?\n/, "");
  const headings = body.split(/\r?\n/)
    .filter((line) => line.startsWith("## "))
    .map((line) => line.slice(3).trim())
    .filter(Boolean);

  const result = {
    ...skill,
    route: `/skills/${skill.name}/`,
    markdownRoute: `/skills/${skill.name}.md`,
    collectionSlug: collection.slug,
    collectionDisplayName: collection.display_name,
    collectionLogoPath: collection.logoPath,
    defaultPrompt: metadata.default_prompt,
    complements: metadata.relations?.complements ?? [],
    overlaps: metadata.relations?.overlaps ?? [],
    headings,
    sourceUrl: `https://github.com/ashwingopalsamy/golangskills.com/blob/main/${skill.path}/SKILL.md`,
    metadataUrl: `https://github.com/ashwingopalsamy/golangskills.com/blob/main/${skill.path}/skill.json`,
    evaluationsUrl: `https://github.com/ashwingopalsamy/golangskills.com/blob/main/${skill.path}/evals.json`,
  };
  collection.skills.push(result);
  return result;
});

export function getSiteData() {
  return { catalog, collections, skills };
}

export function collectionBySlug(slug) {
  return collections.find((collection) => collection.slug === slug);
}

export function skillByName(name) {
  return skills.find((skill) => skill.name === name);
}

export function collectionSlugByName(name) {
  return collectionSlugs.get(name);
}
