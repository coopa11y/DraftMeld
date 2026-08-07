import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

const repositoryRoot = fileURLToPath(new URL("../", import.meta.url));
const version = readFileSync(`${repositoryRoot}/VERSION`, "utf8").trim();
const changelog = readFileSync(`${repositoryRoot}/CHANGELOG.md`, "utf8");
const escapedVersion = version.replaceAll(".", "\\.");
const heading = changelog.match(new RegExp(`^## \\[${escapedVersion}\\].*$`, "m"));

if (heading?.index === undefined) throw new Error(`Unable to find release notes for ${version} in CHANGELOG.md.`);

const contentStart = heading.index + heading[0].length;
const remainingChangelog = changelog.slice(contentStart).trimStart();
const nextRelease = remainingChangelog.search(/^## \[/m);
const releaseNotes = (nextRelease === -1 ? remainingChangelog : remainingChangelog.slice(0, nextRelease)).trim();

process.stdout.write(
  `DraftMeld ${version} is an early public release. Back up important league data before upgrading.\n\n${releaseNotes}\n`,
);
