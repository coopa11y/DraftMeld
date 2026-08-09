import { readFile } from "node:fs/promises";

const expected = (await readFile(new URL("../VERSION", import.meta.url), "utf8")).trim();
const rootPackage = JSON.parse(await readFile(new URL("../package.json", import.meta.url), "utf8"));
const frontendPackage = JSON.parse(await readFile(new URL("../frontend/package.json", import.meta.url), "utf8"));
const openapi = await readFile(new URL("../contracts/openapi.yaml", import.meta.url), "utf8");

const versions = new Map([
  ["root package", rootPackage.version],
  ["frontend package", frontendPackage.version],
  ["OpenAPI contract", openapi.match(/^  version: (.+)$/m)?.[1]],
]);

const mismatches = [...versions].filter(([, version]) => version !== expected);
if (mismatches.length > 0) {
  for (const [source, version] of mismatches) {
    console.error(`${source} has ${version ?? "no version"}; expected ${expected}`);
  }
  process.exit(1);
}

console.log(`DraftMeld version ${expected} is consistent.`);
