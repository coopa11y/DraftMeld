import { useEffect, useRef, useState, type Dispatch, type FormEvent, type RefObject, type SetStateAction } from "react";
import { importESPNLeagueRules, importLeagueRules } from "../shared/api/leagues";
import type { LeagueRuleImport, LeagueRules } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { FormField } from "../shared/ui/FormField";
import { StatusMessage } from "../shared/ui/StatusMessage";
import { applyLeagueRuleImport } from "./applyLeagueRuleImport";
import { cloneLeagueRules } from "./leagueDefaults";
import type { OnboardingImportReview } from "./onboarding";

interface LeagueRulesImportProps {
  rules: LeagueRules;
  setRules: Dispatch<SetStateAction<LeagueRules>>;
  importReview?: OnboardingImportReview;
  setImportReview: Dispatch<SetStateAction<OnboardingImportReview | undefined>>;
}

type ImportMethod = "espn-url" | "espn-paste" | "espn-file" | "file";

export function LeagueRulesImport({ rules, setRules, importReview, setImportReview }: LeagueRulesImportProps) {
  const [method, setMethod] = useState<ImportMethod>("espn-url");
  const [leagueURL, setLeagueURL] = useState("");
  const [season, setSeason] = useState(new Date().getFullYear());
  const [pastedSettings, setPastedSettings] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const resultsHeading = useRef<HTMLHeadingElement>(null);

  useEffect(() => {
    if (importReview) resultsHeading.current?.focus();
  }, [importReview]);

  async function handleImport(event: FormEvent) {
    event.preventDefault();
    setBusy(true);
    setError("");
    try {
      const imported = await runImport();
      setImportReview({ result: imported, beforeRules: cloneLeagueRules(rules) });
      setRules((current) => applyLeagueRuleImport(current, imported));
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "The league settings could not be imported.");
    } finally {
      setBusy(false);
    }
  }

  async function runImport() {
    if (method === "espn-url") return importESPNLeagueRules(leagueURL, season);
    if (method === "espn-paste") {
      return importLeagueRules(new File([pastedSettings], "espn-settings.txt", { type: "text/plain" }), "espn");
    }
    if (!file) throw new Error("Choose a settings file to import.");
    return importLeagueRules(file, method === "espn-file" ? "espn" : undefined);
  }

  function undoImport() {
    if (!importReview) return;
    setRules(cloneLeagueRules(importReview.beforeRules));
    setImportReview(undefined);
  }

  const canSubmit =
    method === "espn-url"
      ? leagueURL.trim().length > 0
      : method === "espn-paste"
        ? pastedSettings.trim().length > 0
        : file !== null;

  return (
    <section className="league-rules-import" aria-labelledby="league-rules-import-title">
      <h3 id="league-rules-import-title">Import league settings</h3>
      <p>
        DraftMeld applies only recognized settings. You review the results before saving, and you can undo the entire
        import.
      </p>
      <form className="rule-import" onSubmit={handleImport}>
        <fieldset disabled={busy}>
          <legend>Choose how to import</legend>
          <FormField label="Import method" help="Choose one method. Only the fields needed for that method are shown.">
            <select
              value={method}
              onChange={(event) => {
                setMethod(event.target.value as ImportMethod);
                setError("");
                setFile(null);
              }}
            >
              <option value="espn-url">ESPN league URL</option>
              <option value="espn-paste">Paste ESPN settings</option>
              <option value="espn-file">ESPN JSON or PDF file</option>
              <option value="file">Other PDF or CSV file</option>
            </select>
          </FormField>

          {method === "espn-url" ? (
            <ESPNURLFields leagueURL={leagueURL} season={season} setLeagueURL={setLeagueURL} setSeason={setSeason} />
          ) : null}
          {method === "espn-paste" ? (
            <FormField
              label="ESPN settings text"
              help="On ESPN's league settings page, select and copy the visible settings, then paste them here. DraftMeld does not need your ESPN password or cookies."
            >
              <textarea rows={8} value={pastedSettings} onChange={(event) => setPastedSettings(event.target.value)} />
            </FormField>
          ) : null}
          {method === "espn-file" ? (
            <ImportFileField
              label="ESPN settings JSON or PDF"
              help="Use JSON you exported from ESPN, or print the complete ESPN settings page to PDF. Scanned PDFs use DraftMeld's existing local OCR."
              accept=".json,.pdf,application/json,application/pdf"
              setFile={setFile}
            />
          ) : null}
          {method === "file" ? (
            <ImportFileField
              label="League settings PDF or CSV"
              help="Maximum file size: 20 MiB. Scanned PDFs use DraftMeld's existing local OCR when it is available."
              accept=".pdf,.csv,application/pdf,text/csv"
              setFile={setFile}
            />
          ) : null}

          <Button type="submit" variant="primary" disabled={!canSubmit || busy}>
            {busy ? "Reading league settings…" : "Import and review"}
          </Button>
        </fieldset>
      </form>

      {error ? (
        <StatusMessage tone="error">
          {error} {method === "espn-url" ? "Private leagues can use the paste or ESPN PDF method instead." : ""}
        </StatusMessage>
      ) : null}
      {importReview ? (
        <ImportResults headingRef={resultsHeading} result={importReview.result} onUndo={undoImport} />
      ) : null}
    </section>
  );
}

function ESPNURLFields({
  leagueURL,
  season,
  setLeagueURL,
  setSeason,
}: {
  leagueURL: string;
  season: number;
  setLeagueURL: (value: string) => void;
  setSeason: (value: number) => void;
}) {
  return (
    <div className="form-grid">
      <FormField
        label="ESPN league URL"
        help="Paste the address from an ESPN fantasy football league page. This direct method works for publicly readable leagues."
      >
        <input
          type="url"
          required
          placeholder="https://fantasy.espn.com/football/league/settings?leagueId=123456"
          value={leagueURL}
          onChange={(event) => setLeagueURL(event.target.value)}
        />
      </FormField>
      <FormField label="ESPN season" help="Use the season whose settings you want to import.">
        <input
          type="number"
          min="2000"
          max={new Date().getFullYear() + 1}
          required
          value={season}
          onChange={(event) => setSeason(Number(event.target.value))}
        />
      </FormField>
    </div>
  );
}

function ImportFileField({
  label,
  help,
  accept,
  setFile,
}: {
  label: string;
  help: string;
  accept: string;
  setFile: (file: File | null) => void;
}) {
  return (
    <FormField label={label} help={help}>
      <input type="file" accept={accept} onChange={(event) => setFile(event.target.files?.[0] ?? null)} />
    </FormField>
  );
}

function ImportResults({
  headingRef,
  result,
  onUndo,
}: {
  headingRef: RefObject<HTMLHeadingElement | null>;
  result: LeagueRuleImport;
  onUndo: () => void;
}) {
  const total = result.settingMatches.length + result.matches.length;
  const sourceName = formatImportSource(result.fileType);
  return (
    <section className="rule-import-results" aria-labelledby="imported-rules-title">
      <h4 id="imported-rules-title" ref={headingRef} tabIndex={-1}>
        Imported values to review
      </h4>
      <StatusMessage tone="success">
        Applied {total} recognized {total === 1 ? "value" : "values"}: {result.settingMatches.length} league settings
        and {result.matches.length} scoring values.
      </StatusMessage>
      {result.settingMatches.length ? (
        <ImportTable
          label="Imported league settings"
          caption={`League settings recognized from ${sourceName}`}
          rows={result.settingMatches}
          valueHeading="Imported value"
        />
      ) : null}
      {result.matches.length ? (
        <ImportTable
          label="Imported scoring values"
          caption={`Scoring rules recognized from ${sourceName}`}
          rows={result.matches.map((match) => ({ ...match, value: String(match.value) }))}
          valueHeading="Points"
        />
      ) : null}
      {result.warnings.map((warning) => (
        <StatusMessage key={warning}>{warning}</StatusMessage>
      ))}
      <Button onClick={onUndo}>Undo all imported values</Button>
    </section>
  );
}

function formatImportSource(fileType: LeagueRuleImport["fileType"]) {
  const labels: Record<LeagueRuleImport["fileType"], string> = {
    csv: "the CSV file",
    pdf: "the PDF file",
    espn: "ESPN",
    "espn-pdf": "the ESPN PDF",
    "espn-text": "the pasted ESPN settings",
  };
  return labels[fileType];
}

interface ImportTableProps {
  label: string;
  caption: string;
  valueHeading: string;
  rows: Array<{ key: string; label: string; value: string; confidence: string; source: string }>;
}

function ImportTable({ label, caption, valueHeading, rows }: ImportTableProps) {
  return (
    <div className="table-scroll" role="region" aria-label={label} tabIndex={0}>
      <table>
        <caption>{caption}</caption>
        <thead>
          <tr>
            <th scope="col">Setting</th>
            <th scope="col">{valueHeading}</th>
            <th scope="col">Confidence</th>
            <th scope="col">Detected source</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr key={row.key}>
              <th scope="row">{row.label}</th>
              <td>{row.value}</td>
              <td>{row.confidence}</td>
              <td>{row.source}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
