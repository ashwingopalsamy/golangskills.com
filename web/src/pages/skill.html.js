import { installationGuideHTML } from "../lib/install-guide.js";

const title = "Install Go Engineering Skills | Go Engineering Skills";
const description = "Installation layouts and commands for Go Engineering Skills, rendered from the canonical docs/install.md guide.";
const canonical = "https://golangskills.com/skill.html";
const image = "https://golangskills.com/assets/engineering-logo.png";
const jsonLd = JSON.stringify({
  "@context": "https://schema.org",
  "@type": "WebPage",
  "@id": canonical,
  url: canonical,
  name: "Install Go Engineering Skills",
  description: "Installation layouts and commands from the canonical repository guide.",
  isPartOf: { "@id": "https://golangskills.com/#website" },
}).replace(/</g, "\\u003c");

export function GET() {
  const guide = installationGuideHTML();
  const html = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="color-scheme" content="light dark">
    <meta name="theme-color" content="#fdfdfd" media="(prefers-color-scheme: light)">
    <meta name="theme-color" content="#070707" media="(prefers-color-scheme: dark)">
    <title>${title}</title>
    <meta name="description" content="${description}">
    <meta name="robots" content="index,follow,max-image-preview:large,max-snippet:-1">
    <link rel="canonical" href="${canonical}">
    <link rel="icon" type="image/png" href="./assets/engineering-logo.png">
    <link rel="stylesheet" href="./assets/site.css">
    <meta property="og:type" content="website">
    <meta property="og:site_name" content="Go Engineering Skills">
    <meta property="og:title" content="${title}">
    <meta property="og:description" content="${description}">
    <meta property="og:url" content="${canonical}">
    <meta property="og:image" content="${image}">
    <meta property="og:image:alt" content="Go Engineering Skills mark">
    <meta name="twitter:card" content="summary">
    <meta name="twitter:title" content="${title}">
    <meta name="twitter:description" content="${description}">
    <meta name="twitter:image" content="${image}">
    <script type="application/ld+json">${jsonLd}</script>
    <script src="./assets/site.js" defer></script>
  </head>
  <body>
    <a class="skip-link" href="#main">Skip to content</a>
    <header class="site-nav">
      <div class="site-nav-inner page-width">
        <div class="site-nav-left">
          <a class="brand" href="./index.html" aria-label="Go Engineering Skills home">
            <img src="./assets/engineering-logo.png" width="24" height="24" alt="">
            <span class="brand-word"><strong>Go skills</strong><span class="brand-word-dim">.com</span></span>
          </a>
          <nav class="site-nav-menu" aria-label="Main navigation">
            <a class="nav-pill" href="./index.html#catalog">Catalog</a>
            <a class="nav-pill" href="./skill.html" aria-current="page">Install guide</a>
            <a class="nav-pill nav-agent-link" href="./llms.txt">For agents</a>
          </nav>
        </div>
        <div class="site-nav-right">
          <a class="utility-pill search-trigger" href="./index.html#catalog" aria-label="Search the skill catalog, Command or Control K">
            <span aria-hidden="true">⌕</span><span class="utility-label">Find a skill</span><kbd>⌘ K</kbd>
          </a>
          <button class="utility-pill theme-toggle enhanced-only" type="button" hidden aria-label="Switch to dark theme" aria-pressed="false">
            <span class="theme-glyph" aria-hidden="true">◐</span><span class="utility-label theme-label">Theme</span>
          </button>
        </div>
      </div>
    </header>
    <main id="main">
      <article class="install-document page-width">
        <nav class="breadcrumbs" aria-label="Breadcrumb"><a href="./index.html">Catalog</a><span aria-hidden="true">/</span><span aria-current="page">Install guide</span></nav>
        <p class="guide-source">Canonical guide: <a href="https://github.com/ashwingopalsamy/golangskills.com/blob/main/docs/install.md">docs/install.md <span aria-hidden="true">↗</span></a></p>
        <div class="install-document-body">${guide}</div>
      </article>
    </main>
    <p class="visually-hidden" data-copy-status role="status" aria-live="polite"></p>
    <footer class="site-footer page-width">
      <a class="brand footer-brand" href="./index.html" aria-label="Go Engineering Skills home">
        <img src="./assets/engineering-logo.png" width="24" height="24" alt="">
        <span class="brand-word"><strong>Go skills</strong><span class="brand-word-dim">.com</span></span>
      </a>
      <nav class="footer-links" aria-label="Project resources">
        <a href="./skill.html">Install guide</a><a href="./sitemap.md">Sitemap</a><a href="./llms.txt">LLM context</a>
        <a href="./.well-known/agents.md">Agent guide</a><a href="https://github.com/ashwingopalsamy/golangskills.com">GitHub source</a>
        <a href="https://github.com/ashwingopalsamy/golangskills.com/blob/main/LICENSE">Apache-2.0</a>
      </nav>
      <p class="footer-note">Independent project. Not endorsed by Google or the Go project.</p>
    </footer>
  </body>
</html>`;
  return new Response(html, { headers: { "Content-Type": "text/html; charset=utf-8" } });
}
