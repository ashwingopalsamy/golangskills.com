import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { repositoryRoot } from "./catalog.js";

const source = readFileSync(resolve(repositoryRoot, "docs/install.md"), "utf8");
const repositoryURL = "https://github.com/ashwingopalsamy/golangskills.com/blob/main";

function escapeHTML(value) {
  return value.replaceAll("&", "&amp;").replaceAll("<", "&lt;").replaceAll(">", "&gt;").replaceAll('"', "&quot;").replaceAll("'", "&#39;");
}

function safeLink(target) {
  if (/^https?:\/\//i.test(target)) return target;
  if (/^(?:\.\/)?[\w./-]+\.md$/.test(target)) {
    const path = target.replace(/^\.\//, "").replace(/^\.\.\//, "");
    return `${repositoryURL}/docs/${path}`;
  }
  return "#";
}

function inlineMarkdown(value) {
  const token = /`([^`]+)`|\[([^\]]+)\]\(([^)]+)\)/g;
  let result = "";
  let previous = 0;
  for (const match of value.matchAll(token)) {
    result += escapeHTML(value.slice(previous, match.index));
    if (match[1]) result += `<code>${escapeHTML(match[1])}</code>`;
    else result += `<a href="${escapeHTML(safeLink(match[3]))}">${escapeHTML(match[2])}</a>`;
    previous = match.index + match[0].length;
  }
  return result + escapeHTML(value.slice(previous));
}

function tableCells(line) {
  return line.trim().replace(/^\|/, "").replace(/\|$/, "").split("|").map((cell) => cell.trim());
}

export function installationGuideHTML() {
  const output = [];
  let paragraph = [];
  let listOpen = false;
  let tableOpen = false;
  let tableHeadOpen = false;
  let codeFence = false;
  let codeLines = [];

  const flushParagraph = () => {
    if (paragraph.length) output.push(`<p>${inlineMarkdown(paragraph.join(" "))}</p>`);
    paragraph = [];
  };
  const closeList = () => {
    if (listOpen) output.push("</ul>");
    listOpen = false;
  };
  const closeTable = () => {
    if (!tableOpen) return;
    output.push("</tbody></table>");
    tableOpen = false;
    tableHeadOpen = false;
  };
  const closeBlocks = () => {
    flushParagraph();
    closeList();
    closeTable();
  };

  for (const line of source.replace(/\r/g, "").split("\n")) {
    if (line.startsWith("```")) {
      if (codeFence) {
        output.push(`<pre><code>${escapeHTML(codeLines.join("\n"))}</code></pre>`);
        codeLines = [];
        codeFence = false;
      } else {
        closeBlocks();
        codeFence = true;
      }
      continue;
    }
    if (codeFence) {
      codeLines.push(line);
      continue;
    }
    const heading = /^(#{1,4})\s+(.+)$/.exec(line);
    if (heading) {
      closeBlocks();
      const level = heading[1].length;
      output.push(`<h${level}>${inlineMarkdown(heading[2])}</h${level}>`);
      continue;
    }
    if (/^\|.*\|$/.test(line.trim())) {
      flushParagraph();
      closeList();
      const cells = tableCells(line);
      if (!tableOpen) {
        output.push("<table><thead><tr>", ...cells.map((cell) => `<th scope=\"col\">${inlineMarkdown(cell)}</th>`), "</tr></thead><tbody>");
        tableOpen = true;
        tableHeadOpen = true;
      } else if (cells.every((cell) => /^:?-{3,}:?$/.test(cell))) {
        tableHeadOpen = false;
      } else {
        if (tableHeadOpen) tableHeadOpen = false;
        output.push("<tr>", ...cells.map((cell) => `<td>${inlineMarkdown(cell)}</td>`), "</tr>");
      }
      continue;
    }
    closeTable();
    if (/^\s*-\s+/.test(line)) {
      flushParagraph();
      if (!listOpen) {
        output.push("<ul>");
        listOpen = true;
      }
      output.push(`<li>${inlineMarkdown(line.replace(/^\s*-\s+/, ""))}</li>`);
      continue;
    }
    if (!line.trim()) {
      flushParagraph();
      closeList();
      continue;
    }
    closeList();
    paragraph.push(line.trim());
  }
  closeBlocks();
  if (codeFence) output.push(`<pre><code>${escapeHTML(codeLines.join("\n"))}</code></pre>`);
  return output.join("\n");
}
