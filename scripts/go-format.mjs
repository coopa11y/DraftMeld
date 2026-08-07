import { readdirSync } from "node:fs";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const backendRoot = fileURLToPath(new URL("../backend/", import.meta.url));
const scriptsRoot = fileURLToPath(new URL("../scripts/", import.meta.url));
const checkOnly = process.argv.includes("--check");

function goFiles(directory) {
  return readdirSync(directory, { withFileTypes: true })
    .flatMap((entry) => {
      const path = `${directory}/${entry.name}`;
      if (entry.isDirectory()) return goFiles(path);
      return entry.isFile() && entry.name.endsWith(".go") ? [path] : [];
    })
    .sort();
}

const files = [...goFiles(backendRoot), ...goFiles(scriptsRoot)].sort();
const executable = process.env.GOFMT || "gofmt";
const result = spawnSync(executable, [checkOnly ? "-l" : "-w", ...files], { encoding: "utf8" });

if (result.error) {
  console.error(`Unable to run gofmt: ${result.error.message}`);
  process.exit(1);
}

if (result.status !== 0) {
  if (result.stderr) console.error(result.stderr.trim());
  process.exit(result.status ?? 1);
}

const unformatted = result.stdout.trim();
if (checkOnly && unformatted) {
  console.error(`The following Go files need formatting:\n${unformatted}`);
  process.exit(1);
}

console.log(checkOnly ? `${files.length} Go files use gofmt style.` : `Formatted ${files.length} Go files.`);
