const siteOrigin = "https://golangskills.com";
const repositoryURL = "https://github.com/ashwingopalsamy/golangskills.com";

export function fullCatalogText({ catalog, collections, skills }) {
  const out = [
    "# Go Engineering Skills: Full Catalog Context",
    "",
    "This file summarizes the public static catalog. The canonical instructions and supporting references remain in the linked GitHub files. This project does not expose an MCP server, REST API, or write-capable agent endpoint.",
    "",
    `- Website: ${siteOrigin}/`,
    `- Repository: ${repositoryURL}`,
    "- License: Apache-2.0",
    `- Catalog version: ${catalog.collection_version}`,
    `- Skills: ${skills.length}`,
    `- Collections: ${collections.length}`,
    "",
    "## Installation",
    "",
    "```sh",
    "npx skills add ashwingopalsamy/golangskills.com",
    "```",
    "",
    "This is the documented open Skills CLI path. Consult the repository installation guide for current client layouts.",
    `- Installation guide: ${siteOrigin}/skill.html`,
    "",
    "## Collections",
    "",
  ];
  for (const collection of collections) {
    out.push(`### ${collection.display_name}`, "", collection.description, "", `Page: ${siteOrigin}${collection.route}`, `Source guide: ${collection.sourceUrl}`, "");
  }
  out.push("## Skills", "");
  for (const skill of skills) {
    out.push(
      `### ${skill.display_name}`,
      "",
      `- Collection: ${skill.collectionDisplayName}`,
      `- Use when: ${skill.description}`,
      `- Compatibility: ${skill.compatibility}`,
      `- Maturity: ${skill.maturity}`,
      `- Category: ${skill.category}`,
      `- Tags: ${skill.tags.join(", ")}`,
      `- Source records: ${skill.source_count}`,
      `- Evaluation cases: ${skill.evaluation_count}`,
      `- Page: ${siteOrigin}${skill.route}`,
      `- Canonical instructions: ${skill.sourceUrl}`,
      `- Metadata: ${skill.metadataUrl}`,
      `- Evaluation cases: ${skill.evaluationsUrl}`,
      "",
    );
  }
  out.push("## Evidence and status", "", "See the canonical claim ledger, locked source snapshots, evaluation reports, compatibility notes, and benchmark status in the repository. A source count or committed eval case is not itself a measured improvement claim.", "");
  return out.join("\n");
}

export function aiSummary({ catalog, collections, skills }) {
  return {
    catalogVersion: catalog.collection_version,
    name: "Go Engineering Skills by Ashwin Gopalsamy",
    summary: "A static catalog of focused, evidence-backed Go skills for AI coding agents, with production engineering, distributed systems, and fintech collections.",
    counts: { skills: skills.length, collections: collections.length },
    license: "Apache-2.0",
    links: {
      home: `${siteOrigin}/`,
      catalog: `${siteOrigin}/sitemap.md`,
      llms: `${siteOrigin}/llms.txt`,
      llmsFull: `${siteOrigin}/llms-full.txt`,
      agentGuide: `${siteOrigin}/.well-known/agents.md`,
      installationGuide: `${siteOrigin}/skill.html`,
      repository: repositoryURL,
      sourceCatalog: `${repositoryURL}/blob/main/catalog/catalog.json`,
    },
    install: "npx skills add ashwingopalsamy/golangskills.com",
    benchmark: {
      status: catalog.benchmark.status,
      leadershipClaimEligible: catalog.benchmark.leadership_claim_eligible,
      blockingGates: catalog.benchmark.blocking_gates,
      source: `${repositoryURL}/blob/main/${catalog.benchmark.evidence_path}`,
    },
    collections: collections.map((collection) => ({
      name: collection.display_name,
      description: collection.description,
      page: `${siteOrigin}${collection.route}`,
      skills: collection.skills.map((skill) => ({
        name: skill.name,
        title: skill.display_name,
        collection: collection.display_name,
        use_when: skill.description,
        compatibility: skill.compatibility,
        page: `${siteOrigin}${skill.route}`,
        canonical_source: skill.sourceUrl,
      })),
    })),
  };
}

export function agentGuide({ collections, skills }) {
  const out = [
    "# Agent guide for golangskills.com",
    "",
    "This is a public, static, read-only catalog. It describes the published Go skill files and their discovery pages. It is not an agent API: there are no MCP or REST endpoints, credentials, user data, or write operations.",
    "",
    "## Start with the source of truth",
    "",
    `- [Website catalog](${siteOrigin}/) for the complete, no-JavaScript skill index.`,
    `- [Installation guide](${siteOrigin}/skill.html) for the documented client layouts and install commands.`,
    `- [Concise context](${siteOrigin}/llms.txt) for collections and entry points.`,
    `- [Full catalog context](${siteOrigin}/llms-full.txt) for machine-readable summaries of every skill.`,
    `- [JSON summary](${siteOrigin}/api/ai-summary.json) for the collections, install command, and current benchmark status.`,
    `- [Discovery catalog](${siteOrigin}/.well-known/ai-catalog.json) for collection and skill page URLs with representative queries.`,
    `- [Markdown sitemap](${siteOrigin}/sitemap.md) and [XML sitemap](${siteOrigin}/sitemap.xml) for public HTML pages.`,
    `- [Canonical repository](${repositoryURL}) for the authored skill corpus and evidence.`,
    "",
    "## Select a skill",
    "",
    "Choose by the task boundary in each skill’s description. Read the canonical `SKILL.md` before relying on a summary; check its one-hop references when that skill points to them. `skill.json` and `evals.json` are also canonical repository inputs. The website mirrors are generated from those files.",
    "",
  ];
  for (const collection of collections) out.push(`- [${collection.display_name}](${siteOrigin}${collection.route}): ${collection.description}`);
  out.push(
    "",
    `The catalog currently exposes ${skills.length} skills across ${collections.length} collections.`,
    "",
    "## Install",
    "",
    "```sh",
    "npx skills add ashwingopalsamy/golangskills.com",
    "```",
    "",
    "This is the documented open Skills CLI installation path. Consult the repository’s installation guide for client layouts and current compatibility evidence. Do not treat a structural layout check as proof of behavioral performance.",
    "",
    `## Boundaries and evidence\n\nSkills are instruction files for coding agents. They do not execute code, sandbox an agent, or guarantee correct output. Claims, source authority, compatibility status, and benchmark limits are tracked in the canonical repository: [claim ledger](${repositoryURL}/blob/main/knowledge/claims/canonical.json), [compatibility notes](${repositoryURL}/blob/main/docs/compatibility.md), and [benchmark status](${repositoryURL}/blob/main/catalog/benchmark-status.json).`,
    "",
    "All pages and files exposed by this site are public static material. Use the canonical GitHub source when a mirrored summary conflicts with a skill file.",
    "",
  );
  return out.join("\n");
}

export function agentCatalog({ collections, skills }) {
  const entries = [
    {
      identifier: `${siteOrigin}/llms.txt`,
      displayName: "Go Engineering Skills concise context",
      type: "text/plain",
      url: `${siteOrigin}/llms.txt`,
      representativeQueries: ["what Go engineering skill collections are available", "how to install the Go skills"],
    },
    {
      identifier: `${siteOrigin}/api/ai-summary.json`,
      displayName: "Go Engineering Skills catalog summary",
      type: "application/json",
      url: `${siteOrigin}/api/ai-summary.json`,
      representativeQueries: ["summarize Go skills and compatibility", "find canonical Go skill sources"],
    },
    {
      identifier: `${siteOrigin}/.well-known/agents.md`,
      displayName: "Go Engineering Skills agent guide",
      type: "text/markdown",
      url: `${siteOrigin}/.well-known/agents.md`,
      representativeQueries: ["how should an agent use this static skills catalog", "what are the source and capability boundaries"],
    },
    {
      identifier: `${siteOrigin}/skill.html`,
      displayName: "Go Engineering Skills installation guide",
      type: "text/html",
      url: `${siteOrigin}/skill.html`,
      representativeQueries: ["how do I install the Go skills", "which Go skill package fits my coding agent"],
    },
  ];
  for (const collection of collections) {
    entries.push({
      identifier: `${siteOrigin}${collection.route}`,
      displayName: collection.display_name,
      type: "text/html",
      url: `${siteOrigin}${collection.route}`,
      representativeQueries: [
        `Go skills for ${collection.display_name.replace(/^Go: /, "")}`,
        `which Go skills fit ${collection.display_name.replace(/^Go: /, "")}`,
      ],
    });
  }
  for (const skill of skills) {
    entries.push({
      identifier: `${siteOrigin}${skill.route}`,
      displayName: skill.display_name,
      type: "text/html",
      url: `${siteOrigin}${skill.route}`,
      representativeQueries: [skill.description, `Go skill for ${skill.display_name.toLowerCase()}`],
    });
  }
  return {
    specVersion: "1.0",
    host: { displayName: "Go Engineering Skills", identifier: siteOrigin },
    entries,
  };
}

export function markdownSitemap({ collections, skills }) {
  const out = [
    "# Sitemap",
    "",
    `Public pages on golangskills.com. Agent discovery material is listed in [the public agent guide](${siteOrigin}/.well-known/agents.md).`,
    "",
    "## Catalog",
    "",
    `- [Go Skills for Production Systems](${siteOrigin}/)`,
    "",
    "## Collections",
    "",
  ];
  for (const collection of collections) out.push(`- [${collection.display_name}](${siteOrigin}${collection.route})`);
  out.push("", "## Guides", "", `- [Installation guide](${siteOrigin}/skill.html)`, "", "## Skills", "");
  for (const skill of skills) out.push(`- [${skill.display_name}](${siteOrigin}${skill.route})`);
  out.push("");
  return out.join("\n");
}

export function xmlSitemap({ collections, skills }) {
  const urls = ["/", "/skill.html", ...collections.map((item) => item.route), ...skills.map((item) => item.route)];
  const rows = urls.map((route) => `  <url><loc>${siteOrigin}${route}</loc></url>`);
  return `<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n${rows.join("\n")}\n</urlset>\n`;
}
