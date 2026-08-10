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
  sources: RankingSource[];
  onImport: (name: string, file: File, mapping: Record<string, string>) => Promise<void>;
}

export function RankingCsvImport({ busy, embedded = false, sources, onImport }: RankingCsvImportProps) {
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
    setMapping(autoMapCsvHeaders(parsedHeaders, rankingColumns));
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = event.currentTarget;
    if (!name.trim() || !file || requiredRankingColumns.some((column) => !mapping[column.key])) return;
    try {
      await onImport(name, file, mapping);
    } catch {
      return;
    }
    setName("");
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
          <p className="eyebrow">Your rankings</p>
          <h2 id="ranking-csv-import-heading">Import a ranking CSV</h2>
          <p className="section-description">
            Add a private ordinal ranking from any service or spreadsheet. Confirm the columns before importing; the
            uploaded file is processed locally by DraftMeld and is not retained.
          </p>
          <a className="inline-action-link" href={rankingTemplate} download="draftmeld-ranking-template.csv">
            Download ranking CSV template
          </a>
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
        {headers.length > 0 ? (
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
          disabled={busy || !name.trim() || !file || requiredRankingColumns.some((column) => !mapping[column.key])}
        >
          Import rankings
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
