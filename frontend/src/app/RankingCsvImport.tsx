import { useState, type FormEvent } from "react";
import type { RankingSource } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import {
  autoMapCsvHeaders,
  CsvColumnMapper,
  inspectCsvHeaders,
  type CsvColumnDefinition,
} from "../shared/ui/CsvColumnMapper";
import { FormField } from "../shared/ui/FormField";
import { Panel } from "../shared/ui/Panel";
import { StatusMessage } from "../shared/ui/StatusMessage";

const rankingTemplate =
  "data:text/csv;charset=utf-8,rank%2Cname%2Cposition%2Cteam%2Cadp%2Ctier%2CproviderId%0A1%2C%2CRB%2C%2C%2C%2C";

const rankingColumns: CsvColumnDefinition[] = [
  { key: "name", label: "Player name", required: true, aliases: ["name", "player", "playername", "playerfullname"] },
  { key: "rank", label: "Overall rank", required: true, aliases: ["rank", "overallrank", "rk"] },
  { key: "position", label: "Position", required: true, aliases: ["position", "pos"] },
  { key: "team", label: "NFL team", aliases: ["team", "nflteam", "tm"] },
  { key: "adp", label: "Average draft position", aliases: ["adp", "averagedraftposition"] },
  { key: "tier", label: "Tier" },
  { key: "providerId", label: "Source player ID", aliases: ["providerid", "playerid", "id"] },
];
const requiredRankingColumns = rankingColumns.filter((column) => column.required);

interface RankingCsvImportProps {
  embedded?: boolean;
  busy: boolean;
  provider?: "generic" | "udk";
  sources: RankingSource[];
  onImport: (name: string, file: File, mapping: Record<string, string>) => Promise<void>;
}

export function RankingCsvImport({
  busy,
  embedded = false,
  provider = "generic",
  sources,
  onImport,
}: RankingCsvImportProps) {
  const isUDK = provider === "udk";
  const [name, setName] = useState(isUDK ? "Fantasy Footballers UDK Top 200" : "");
  const [file, setFile] = useState<File | null>(null);
  const [headers, setHeaders] = useState<string[]>([]);
  const [mapping, setMapping] = useState<Record<string, string>>({});
  const detectedFormat = detectRankingCsvFormat(headers);
  const fileIsValid = !isUDK || detectedFormat === "udk-top-200";

  async function chooseFile(selected: File | null) {
    setFile(selected);
    if (!selected) {
      setHeaders([]);
      setMapping({});
      return;
    }
    const parsedHeaders = await inspectCsvHeaders(selected);
    setHeaders(parsedHeaders);
    setMapping(autoMapCsvHeaders(parsedHeaders, rankingColumns));
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = event.currentTarget;
    if (!name.trim() || !file || !fileIsValid || requiredRankingColumns.some((column) => !mapping[column.key])) return;
    try {
      await onImport(name, file, mapping);
    } catch {
      return;
    }
    setName(isUDK ? "Fantasy Footballers UDK Top 200" : "");
    setFile(null);
    setHeaders([]);
    setMapping({});
    form.reset();
  }

  const customSources = sources.filter((source) => source.isCustom);

  const content = (
    <>
      <div className="section-heading">
        <div>
          <p className="eyebrow">{isUDK ? "Subscriber rankings" : "Your rankings"}</p>
          <h2 id="ranking-csv-import-heading">{isUDK ? "Import Fantasy Footballers UDK" : "Import a ranking CSV"}</h2>
          {isUDK ? (
            <>
              <p className="section-description">
                Import the Top 200 export from your UDK account. DraftMeld reads it locally and never stores your UDK
                login, browser session, or original file.
              </p>
              <ol className="compact-instructions">
                <li>
                  Open the{" "}
                  <a
                    href="https://www.thefantasyfootballers.com/2026-ultimate-draft-kit/udk-top-200-list/"
                    target="_blank"
                    rel="noreferrer"
                  >
                    UDK Top 200
                  </a>
                  , then choose your scoring settings.
                </li>
                <li>Choose More, then Download CSV.</li>
                <li>Return here and select that Top 200 file.</li>
              </ol>
            </>
          ) : (
            <>
              <p className="section-description">
                Add a private ordinal ranking from any service or spreadsheet. Confirm the columns before importing; the
                uploaded file is processed locally by DraftMeld and is not retained.
              </p>
              <a className="inline-action-link" href={rankingTemplate} download="draftmeld-ranking-template.csv">
                Download ranking CSV template
              </a>
            </>
          )}
        </div>
      </div>
      <form className="data-import-form" noValidate onSubmit={submit}>
        <FormField label="Ranking source name">
          <input
            required
            maxLength={80}
            value={name}
            onChange={(event) => setName(event.target.value)}
            placeholder="My draft rankings"
            disabled={busy}
          />
        </FormField>
        <FormField label="Ranking CSV">
          <input
            type="file"
            accept="text/csv,.csv"
            required
            onChange={(event) => void chooseFile(event.target.files?.[0] ?? null)}
            disabled={busy}
          />
        </FormField>
        {isUDK && detectedFormat === "udk-top-200" ? (
          <StatusMessage>
            UDK Top 200 recognized. Overall rank, player, position, and team are ready to import.
          </StatusMessage>
        ) : null}
        {isUDK && detectedFormat === "udk-position" ? (
          <StatusMessage tone="error">
            This is a position-ranking export. Download the UDK Top 200 CSV so DraftMeld can preserve the published
            overall order.
          </StatusMessage>
        ) : null}
        {isUDK && headers.length > 0 && detectedFormat === "unknown" ? (
          <StatusMessage tone="error">This file is not recognized as a UDK Top 200 export.</StatusMessage>
        ) : null}
        {!isUDK && headers.length > 0 ? (
          <CsvColumnMapper
            busy={busy}
            columns={rankingColumns}
            headers={headers}
            mapping={mapping}
            optionalLabel="Map optional ranking details"
            onChange={setMapping}
          />
        ) : null}
        <Button
          type="submit"
          disabled={
            busy ||
            !name.trim() ||
            !file ||
            !fileIsValid ||
            requiredRankingColumns.some((column) => !mapping[column.key])
          }
        >
          {isUDK ? "Import UDK rankings" : "Import rankings"}
        </Button>
      </form>
      {customSources.length > 0 ? (
        <ul className="compact-data-list" aria-label="Private ranking sources">
          {customSources.map((source) => (
            <li key={source.id}>
              <strong>{source.name}</strong>
              <span>{source.recordCount} players</span>
            </li>
          ))}
        </ul>
      ) : (
        <p className="empty-state panel-empty-state">No private ranking CSVs imported yet.</p>
      )}
    </>
  );
  return embedded ? <div className="embedded-import">{content}</div> : <Panel variant="ranking">{content}</Panel>;
}

type RankingCsvFormat = "udk-top-200" | "udk-position" | "unknown";

function detectRankingCsvFormat(headers: string[]): RankingCsvFormat {
  const normalized = new Set(headers.map((header) => header.toLowerCase().replace(/[^a-z0-9]/g, "")));
  const includes = (...required: string[]) => required.every((header) => normalized.has(header));
  if (includes("name", "position", "team", "byeweek", "rank", "points", "risk", "upside", "adp", "tier", "outlook")) {
    return "udk-position";
  }
  if (includes("rank", "name", "bye", "team", "pos", "andy", "jason", "mike", "markers")) {
    return "udk-top-200";
  }
  return "unknown";
}
