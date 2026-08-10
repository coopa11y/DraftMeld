import { useState, type FormEvent } from "react";
import { importProjectionCSV } from "../shared/api/rankings";
import type { ProjectionSource } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import {
  autoMapCsvHeaders,
  CsvColumnMapper,
  inspectCsvHeaders,
  type CsvColumnDefinition,
} from "../shared/ui/CsvColumnMapper";
import { FormField } from "../shared/ui/FormField";
import { Panel } from "../shared/ui/Panel";
import { projectionScoringFields } from "./scoring";

const projectionTemplate = `data:text/csv;charset=utf-8,${encodeURIComponent(
  [
    "name",
    "position",
    "team",
    "adp",
    "byeWeek",
    ...projectionScoringFields.map((field) => field.key),
    "providerId",
  ].join(",") + "\n",
)}`;

const projectionColumns: CsvColumnDefinition[] = [
  { key: "name", label: "Player name", required: true, aliases: ["name", "player", "playername", "playerfullname"] },
  { key: "position", label: "Position", required: true, aliases: ["position", "pos"] },
  { key: "team", label: "NFL team", required: true, aliases: ["team", "nflteam", "tm"] },
  { key: "adp", label: "Average draft position", aliases: ["adp", "averagedraftposition"] },
  { key: "byeWeek", label: "Bye week", aliases: ["bye", "byeweek"] },
  { key: "providerId", label: "Source player ID", aliases: ["providerid", "playerid", "id"] },
  ...projectionScoringFields.map((field) => ({
    key: field.key,
    label: field.projectionLabel ?? field.label,
    aliases: [field.key, ...(field.aliases ?? [])],
  })),
];
const requiredProjectionColumns = projectionColumns.filter((column) => column.required);

interface ProjectionImportProps {
  embedded?: boolean;
  busy: boolean;
  sources: ProjectionSource[];
  onBusyChange: (busy: boolean) => void;
  onImported: (source: ProjectionSource) => void;
  onMessage: (message: string) => void;
  onError: (message: string) => void;
}

export function ProjectionImport({
  busy,
  embedded = false,
  sources,
  onBusyChange,
  onImported,
  onMessage,
  onError,
}: ProjectionImportProps) {
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
    const parsedHeaders = await inspectCsvHeaders(selected);
    setHeaders(parsedHeaders);
    setMapping(autoMapCsvHeaders(parsedHeaders, projectionColumns));
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = event.currentTarget;
    if (!name.trim() || !file || requiredProjectionColumns.some((column) => !mapping[column.key])) return;
    onBusyChange(true);
    onError("");
    try {
      const imported = await importProjectionCSV(name, file, mapping);
      onImported(imported);
      onMessage(`${imported.name} imported with ${imported.recordCount} granular player projections.`);
      setName("");
      setFile(null);
      setHeaders([]);
      setMapping({});
      form.reset();
    } catch (reason) {
      onError(reason instanceof Error ? reason.message : "Unable to import projection CSV.");
    } finally {
      onBusyChange(false);
    }
  }

  const content = (
    <>
      <div className="section-heading">
        <div>
          <p className="eyebrow">League scoring</p>
          <h2 id="projection-import-heading">Projection sources</h2>
          <p className="section-description">
            Choose any projection CSV, then confirm which columns contain the player identity and statistics. DraftMeld
            uses the league scoring rules to calculate points, replacement value, tiers, and auction values.
          </p>
          <a className="inline-action-link" href={projectionTemplate} download="draftmeld-projection-template.csv">
            Download projection CSV template
          </a>
        </div>
      </div>
      <form className="data-import-form" noValidate onSubmit={submit}>
        <FormField label="Projection source name">
          <input
            required
            value={name}
            onChange={(event) => setName(event.target.value)}
            placeholder="My projection model"
            disabled={busy}
          />
        </FormField>
        <FormField label="Projection CSV">
          <input
            type="file"
            accept="text/csv,.csv"
            required
            onChange={(event) => void chooseFile(event.target.files?.[0] ?? null)}
            disabled={busy}
          />
        </FormField>
        {headers.length > 0 ? (
          <CsvColumnMapper
            busy={busy}
            columns={projectionColumns}
            headers={headers}
            mapping={mapping}
            optionalLabel="Map optional scoring columns"
            onChange={setMapping}
          />
        ) : null}
        <Button
          type="submit"
          disabled={busy || !name.trim() || !file || requiredProjectionColumns.some((column) => !mapping[column.key])}
        >
          Import projections
        </Button>
      </form>
      {sources.length > 0 ? (
        <ul className="compact-data-list">
          {sources.map((source) => (
            <li key={source.id}>
              <strong>{source.name}</strong>
              <span>{source.recordCount} players</span>
            </li>
          ))}
        </ul>
      ) : (
        <p className="empty-state panel-empty-state">
          No projections imported. Consensus rankings still work, but projected points and VOR remain unavailable.
        </p>
      )}
    </>
  );
  return embedded ? <div className="embedded-import">{content}</div> : <Panel variant="ranking">{content}</Panel>;
}
