import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

const repositoryRoot = fileURLToPath(new URL("../", import.meta.url));
const version = readFileSync(`${repositoryRoot}/VERSION`, "utf8").trim();
const changelog = readFileSync(`${repositoryRoot}/CHANGELOG.md`, "utf8");
const suppliedTag = process.argv[2];

const semanticVersion =
  /^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$/;
if (!semanticVersion.test(version)) {
  throw new Error(`VERSION must contain a valid semantic version; received ${JSON.stringify(version)}.`);
}

const escapedVersion = version.replaceAll(".", "\\.");
const releaseHeading = new RegExp(`^## \\[${escapedVersion}\\] - \\d{4}-\\d{2}-\\d{2}$`, "m");
if (!releaseHeading.test(changelog)) {
  throw new Error(`CHANGELOG.md must contain a dated ## [${version}] release heading.`);
}

if (suppliedTag && suppliedTag !== `v${version}`) {
  throw new Error(`Release tag ${suppliedTag} does not match VERSION ${version}. Expected v${version}.`);
}

console.log(`DraftMeld release ${version}${suppliedTag ? ` matches tag ${suppliedTag}` : " is prepared"}.`);
