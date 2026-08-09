import { rmSync } from "node:fs";
import { spawnSync } from "node:child_process";
import path from "node:path";
import process from "node:process";

const repositoryRoot = path.resolve(import.meta.dirname, "..");
const backendDirectory = path.join(repositoryRoot, "backend");
const profile = path.join(backendDirectory, ".coverage.out");
const minimum = 67;

function runGo(arguments_, options = {}) {
  const result = spawnSync(process.platform === "win32" ? "go.exe" : "go", arguments_, {
    cwd: backendDirectory,
    encoding: "utf8",
    stdio: options.capture ? ["ignore", "pipe", "inherit"] : "inherit",
  });
  if (result.status !== 0) process.exit(result.status ?? 1);
  return result.stdout ?? "";
}

try {
  runGo(["test", "./...", `-coverprofile=${profile}`]);
  const report = runGo(["tool", "cover", `-func=${profile}`], { capture: true });
  const match = report.match(/total:\s+\(statements\)\s+([\d.]+)%/);
  if (!match) throw new Error("Go did not report total statement coverage.");
  const actual = Number(match[1]);
  console.log(`Backend statement coverage: ${actual.toFixed(1)}% (minimum ${minimum.toFixed(1)}%).`);
  if (actual < minimum) process.exitCode = 1;
} finally {
  rmSync(profile, { force: true });
}
