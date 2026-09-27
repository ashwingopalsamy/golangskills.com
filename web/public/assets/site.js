(() => {
  const root = document.documentElement;
  const cards = Array.from(document.querySelectorAll("[data-skill-card]"));
  const grid = document.querySelector("[data-skill-grid]");
  const filterButtons = Array.from(document.querySelectorAll("[data-filter]"));
  const searchInput = document.getElementById("skill-search");
  const sortSelect = document.getElementById("skill-sort");
  const noResults = document.querySelector("[data-no-results]");
  const resultCount = document.querySelector("[data-result-count]");
  const copyStatus = document.querySelector("[data-copy-status]");
  const themeButton = document.querySelector(".theme-toggle");
  const searchTrigger = document.querySelector(".search-trigger");
  const collator = new Intl.Collator(undefined, { sensitivity: "base", numeric: true });
  let activeCollection = "all";

  function readTheme() {
    try {
      const stored = localStorage.getItem("golangskills-theme");
      return stored === "light" || stored === "dark" ? stored : null;
    } catch {
      return null;
    }
  }

  function systemTheme() {
    return window.matchMedia?.("(prefers-color-scheme: dark)").matches ? "dark" : "light";
  }

  function updateThemeControl() {
    if (!themeButton) return;
    const theme = root.dataset.theme || systemTheme();
    themeButton.setAttribute("aria-pressed", String(theme === "dark"));
    themeButton.setAttribute("aria-label", `Switch to ${theme === "dark" ? "light" : "dark"} theme`);
    const label = themeButton.querySelector(".theme-label");
    if (label) label.textContent = theme === "dark" ? "Light" : "Dark";
    const glyph = themeButton.querySelector(".theme-glyph");
    if (glyph) glyph.textContent = theme === "dark" ? "☼" : "◐";
  }

  const savedTheme = readTheme();
  if (savedTheme) root.dataset.theme = savedTheme;

  function normalize(value) {
    return value.normalize("NFKD").toLocaleLowerCase().trim();
  }

  function applyCatalog() {
    if (!grid) return;
    const query = normalize(searchInput?.value || "");
    const visible = [];
    for (const card of cards) {
      const matchesCollection = activeCollection === "all" || card.dataset.collection === activeCollection;
      const matchesQuery = !query || normalize(card.dataset.search || "").includes(query);
      card.hidden = !(matchesCollection && matchesQuery);
      if (!card.hidden) visible.push(card);
    }
    if (noResults) noResults.hidden = visible.length > 0;
    if (resultCount) resultCount.textContent = `${visible.length} ${visible.length === 1 ? "skill" : "skills"} shown`;
  }

  function sortCatalog() {
    if (!grid || !sortSelect) return;
    const ordered = cards.slice().sort((left, right) => {
      if (sortSelect.value === "az") return collator.compare(left.dataset.name || "", right.dataset.name || "");
      return Number(left.dataset.index) - Number(right.dataset.index);
    });
    for (const card of ordered) grid.append(card);
  }

  async function copyText(value) {
    if (navigator.clipboard?.writeText) {
      try {
        await navigator.clipboard.writeText(value);
        return true;
      } catch {
        // Fall through to the user-initiated selection copy path.
      }
    }
    const field = document.createElement("textarea");
    field.value = value;
    field.setAttribute("readonly", "");
    field.style.position = "fixed";
    field.style.left = "-9999px";
    document.body.append(field);
    field.select();
    let copied = false;
    try {
      copied = document.execCommand("copy");
    } catch {
      copied = false;
    }
    field.remove();
    return copied;
  }

  document.querySelectorAll("[data-copy-prompt]").forEach((button) => {
    button.addEventListener("click", async () => {
      const card = button.closest("[data-skill-card]");
      const prompt = card?.querySelector("[data-prompt-text]")?.textContent?.trim();
      if (!prompt) return;
      const copied = await copyText(prompt);
      const label = button.querySelector("span");
      const originalLabel = button.getAttribute("aria-label");
      if (copied) {
        button.dataset.copied = "true";
        if (label) label.textContent = "Copied";
        button.setAttribute("aria-label", `Copied default prompt for ${card.dataset.name}`);
        if (copyStatus) copyStatus.textContent = `Default prompt copied for ${card.dataset.name}.`;
        window.setTimeout(() => {
          delete button.dataset.copied;
          if (label) label.textContent = "Copy prompt";
          if (originalLabel) button.setAttribute("aria-label", originalLabel);
        }, 1600);
      } else if (copyStatus) {
        copyStatus.textContent = "Copy was unavailable. Select the prompt text to copy it.";
      }
    });
  });

  for (const button of filterButtons) {
    button.addEventListener("click", () => {
      activeCollection = button.dataset.filter || "all";
      for (const item of filterButtons) item.setAttribute("aria-pressed", String(item === button));
      applyCatalog();
    });
  }

  searchInput?.addEventListener("input", applyCatalog);
  sortSelect?.addEventListener("change", () => {
    sortCatalog();
    applyCatalog();
  });

  themeButton?.addEventListener("click", () => {
    const current = root.dataset.theme || systemTheme();
    const next = current === "dark" ? "light" : "dark";
    root.dataset.theme = next;
    try {
      localStorage.setItem("golangskills-theme", next);
    } catch {
      // Theme changes still apply for this page load when storage is unavailable.
    }
    updateThemeControl();
  });
  updateThemeControl();

  searchTrigger?.addEventListener("click", (event) => {
    if (!searchInput) return;
    event.preventDefault();
    searchInput.focus();
  });

  document.addEventListener("keydown", (event) => {
    if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
      event.preventDefault();
      if (searchInput) searchInput.focus();
      else searchTrigger?.click();
    }
    if (event.key === "Escape" && document.activeElement === searchInput && searchInput) {
      if (searchInput.value) {
        searchInput.value = "";
        applyCatalog();
      } else {
        searchInput.blur();
      }
    }
  });

  root.classList.add("js");
  document.querySelectorAll(".enhanced-only[hidden]").forEach((control) => { control.hidden = false; });
  if (searchInput) applyCatalog();
})();
