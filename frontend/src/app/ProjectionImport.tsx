import { useState, type FormEvent } from "react";
import { importProjectionCSV } from "../shared/api/rankings";
import type { ProjectionSource } from "../shared/api/types";

const projectionTemplate = "data:text/csv;charset=utf-8,name%2Cposition%2Cteam%2Cadp%2CbyeWeek%2Creception%2CpassingYard%2CpassingTouchdown%2Cinterception%2CrushingYard%2CrushingTouchdown%2CreceivingYard%2CreceivingTouchdown%2CfieldGoalMade%2CextraPointMade%2CdefenseSack%2CdefenseInterception%2CdefenseFumbleRecovery%2CdefenseTouchdown%2CdefenseSafety%0A";

const requiredColumns = [["name", "Player name"], ["position", "Position"], ["team", "NFL team"]] as const;
const optionalColumns = [
  ["adp", "Average draft position"], ["byeWeek", "Bye week"], ["reception", "Receptions"],
  ["passingYard", "Passing yards"], ["passingTouchdown", "Passing touchdowns"], ["interception", "Interceptions"],
  ["rushingYard", "Rushing yards"], ["rushingTouchdown", "Rushing touchdowns"], ["receivingYard", "Receiving yards"],
  ["receivingTouchdown", "Receiving touchdowns"], ["fieldGoalMade", "Field goals made"], ["extraPointMade", "Extra points made"],
  ["defenseSack", "Defensive sacks"], ["defenseInterception", "Defensive interceptions"], ["defenseFumbleRecovery", "Fumble recoveries"],
  ["defenseTouchdown", "Defensive touchdowns"], ["defenseSafety", "Safeties"],
] as const;

const headerAliases: Record<string, string[]> = {
  name: ["name", "player", "playername", "playerfullname"], position: ["position", "pos"], team: ["team", "nflteam", "tm"],
  adp: ["adp", "averagedraftposition"], byeWeek: ["bye", "byeweek"], passingYard: ["passingyard", "passingyards", "passyds"],
  passingTouchdown: ["passingtouchdown", "passingtouchdowns", "passtd"], rushingYard: ["rushingyard", "rushingyards", "rushyds"],
  rushingTouchdown: ["rushingtouchdown", "rushingtouchdowns", "rushtd"], receivingYard: ["receivingyard", "receivingyards", "recyds"],
  receivingTouchdown: ["receivingtouchdown", "receivingtouchdowns", "rectd"], reception: ["reception", "receptions", "rec"],
};

interface ProjectionImportProps {
  busy: boolean;
  sources: ProjectionSource[];
  onBusyChange: (busy: boolean) => void;
  onImported: (source: ProjectionSource) => void;
  onMessage: (message: string) => void;
  onError: (message: string) => void;
}

export function ProjectionImport({ busy, sources, onBusyChange, onImported, onMessage, onError }: ProjectionImportProps) {
  const [name, setName] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [headers, setHeaders] = useState<string[]>([]);
  const [mapping, setMapping] = useState<Record<string, string>>({});

  async function chooseFile(selected: File | null) {
    setFile(selected);
    if (!selected) {
      setHeaders([]);
      setMapping({});
      return;
    }
    const parsedHeaders = parseCsvHeader((await selected.slice(0, 16_384).text()).split(/\r?\n/, 1)[0] ?? "");
    setHeaders(parsedHeaders);
    setMapping(autoMapHeaders(parsedHeaders));
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = event.currentTarget;
    if (!name.trim() || !file || requiredColumns.some(([column]) => !mapping[column])) return;
    onBusyChange(true);
    onError("");
    try {
      const imported = await importProjectionCSV(name, file, mapping);
      onImported(imported);
      onMessage(`${imported.name} imported with ${imported.recordCount} granular player projections.`);
      setName(""); setFile(null); setHeaders([]); setMapping({});
      form.reset();
    } catch (reason) {
      onError(reason instanceof Error ? reason.message : "Unable to import projection CSV.");
    } finally {
      onBusyChange(false);
    }
  }

  const mappingControl = ([column, label]: readonly [string, string], required = false) => (
    <label key={column}>{label}<select required={required} value={mapping[column] ?? ""} onChange={(event) => setMapping((current) => ({ ...current, [column]: event.target.value }))} disabled={busy}>
      <option value="">{required ? "Choose a CSV column" : "Not included"}</option>
      {headers.map((header) => <option key={header} value={header}>{header}</option>)}
    </select></label>
  );

  return (
    <section className="ranking-panel" aria-labelledby="projection-import-heading">
      <div className="section-heading"><div><p className="eyebrow">League scoring</p><h2 id="projection-import-heading">Projection sources</h2><p className="section-description">Choose any projection CSV, then confirm which columns contain the player identity and statistics. DraftMeld uses the league scoring rules to calculate points, replacement value, tiers, and auction values.</p><a className="inline-action-link" href={projectionTemplate} download="draftmeld-projection-template.csv">Download projection CSV template</a></div></div>
      <form className="data-import-form" noValidate onSubmit={submit}>
        <label>Projection source name<input required value={name} onChange={(event) => setName(event.target.value)} placeholder="My projection model" disabled={busy} /></label>
        <label>Projection CSV<input type="file" accept="text/csv,.csv" required onChange={(event) => void chooseFile(event.target.files?.[0] ?? null)} disabled={busy} /></label>
        {headers.length > 0 ? <fieldset className="column-mapper"><legend>Match required columns</legend>{requiredColumns.map((field) => mappingControl(field, true))}<details><summary>Map optional scoring columns</summary><div className="column-mapper-grid">{optionalColumns.map((field) => mappingControl(field))}</div></details></fieldset> : null}
        <button className="secondary-button" type="submit" disabled={busy || !name.trim() || !file || requiredColumns.some(([column]) => !mapping[column])}>Import projections</button>
      </form>
      {sources.length > 0 ? <ul className="compact-data-list">{sources.map((source) => <li key={source.id}><strong>{source.name}</strong><span>{source.recordCount} players</span></li>)}</ul> : <p className="empty-state panel-empty-state">No projections imported. Consensus rankings still work, but projected points and VOR remain unavailable.</p>}
    </section>
  );
}

function autoMapHeaders(headers: string[]): Record<string, string> {
  const normalized = new Map(headers.map((header) => [normalizeHeader(header), header]));
  const mapping: Record<string, string> = {};
  for (const [column] of [...requiredColumns, ...optionalColumns]) {
    const aliases = headerAliases[column] ?? [normalizeHeader(column)];
    const match = aliases.map((alias) => normalized.get(alias)).find(Boolean);
    if (match) mapping[column] = match;
  }
  return mapping;
}

function normalizeHeader(value: string): string { return value.trim().toLowerCase().replace(/[^a-z0-9]+/g, ""); }

function parseCsvHeader(line: string): string[] {
  const fields: string[] = [];
  let current = "", quoted = false;
  for (let index = 0; index < line.length; index++) {
    const character = line[index];
    if (character === '"' && quoted && line[index + 1] === '"') { current += '"'; index++; }
    else if (character === '"') quoted = !quoted;
    else if (character === "," && !quoted) { fields.push(current.trim()); current = ""; }
    else current += character;
  }
  fields.push(current.trim());
  return fields.filter(Boolean);
}
