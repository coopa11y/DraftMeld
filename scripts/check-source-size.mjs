import { readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import process from "node:process";

const repositoryRoot = path.resolve(import.meta.dirname, "..");
const policies = [
  { directory: "backend", extensions: new Set([".go"]), maximum: 500 },
  { directory: "frontend/src", extensions: new Set([".ts", ".tsx"]), maximum: 425 },
  { directory: "apps/web/src", extensions: new Set([".ts", ".tsx"]), maximum: 300 },
];

function filesBelow(directory) {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const resolved = path.join(directory, entry.name);
    return entry.isDirectory() ? filesBelow(resolved) : [resolved];
  });
}

const violations = [];
for (const policy of policies) {
  const root = path.join(repositoryRoot, policy.directory);
  for (const file of filesBelow(root)) {
    const relative = path.relative(repositoryRoot, file).replaceAll(path.sep, "/");
    if (
      !policy.extensions.has(path.extname(file)) ||
      relative.includes("/generated.") ||
      relative.match(/(_test|\.test)\.(go|tsx?)$/)
    )
      continue;
    const lines = readFileSync(file, "utf8").split(/\r?\n/).length;
    if (lines > policy.maximum) violations.push(`${relative}: ${lines} lines (maximum ${policy.maximum})`);
  }
}

if (violations.length) {
  console.error(
    "Source files exceed the maintainability limit:\n" + violations.map((value) => `- ${value}`).join("\n"),
  );
  process.exit(1);
}
console.log("Production source files stay within the maintainability limits.");
