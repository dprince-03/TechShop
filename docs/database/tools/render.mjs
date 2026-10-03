// Renders every ```mermaid block in the given Markdown files to diagrams/<file>--<section>.svg.
// Fails (exit 1) if any diagram doesn't render, so it doubles as a syntax check.
// Usage: MMDC=/path/to/mmdc PUPPETEER_CONFIG=/path/to/puppeteer.json node render.mjs out-dir file.md...
import { spawnSync } from "node:child_process";
import { mkdtempSync, readFileSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { basename, join } from "node:path";

const [outDir, ...files] = process.argv.slice(2);
const mmdc = process.env.MMDC ?? "mmdc";
const puppeteer = process.env.PUPPETEER_CONFIG;
const tmp = mkdtempSync(join(tmpdir(), "techshop-mermaid-"));
const slug = (s) => s.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/(^-|-$)/g, "").slice(0, 60);

let ok = 0;
const failed = [];
for (const file of files) {
  const md = readFileSync(file, "utf8");
  const name = basename(file, ".md");
  let heading = "diagram";
  const used = new Map();
  const lines = md.split("\n");
  for (let i = 0; i < lines.length; i++) {
    const h = lines[i].match(/^#{2,3}\s+(.*)/);
    if (h) heading = h[1];
    if (lines[i].trim() !== "```mermaid") continue;
    const end = lines.indexOf("```", i + 1);
    const source = lines.slice(i + 1, end).join("\n");
    i = end;
    let id = `${name}--${slug(heading)}`;
    used.set(id, (used.get(id) ?? 0) + 1);
    if (used.get(id) > 1) id += `-${used.get(id)}`;
    const input = join(tmp, `${id}.mmd`);
    writeFileSync(input, source);
    const args = ["-i", input, "-o", join(outDir, `${id}.svg`), "-b", "white", "-q"];
    if (puppeteer) args.push("-p", puppeteer);
    const r = spawnSync(mmdc, args, { encoding: "utf8" });
    if (r.status === 0) ok++;
    else failed.push(`${id}: ${(r.stderr || r.stdout).split("\n").find((l) => l.trim()) ?? "unknown error"}`);
  }
}
console.log(`Rendered ${ok} diagrams to ${outDir}.`);
if (failed.length) {
  console.error(`FAILED (${failed.length}):\n  ${failed.join("\n  ")}`);
  process.exit(1);
}
