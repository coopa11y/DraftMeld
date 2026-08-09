import { useState, type Dispatch, type FormEvent, type SetStateAction } from "react";
import { importLeagueRules } from "../shared/api/leagues";
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

export function LeagueRulesImport({ rules, setRules, importReview, setImportReview }: LeagueRulesImportProps) {
  const [file, setFile] = useState<File | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  async function handleImport(event: FormEvent) {
    event.preventDefault();
    if (!file) return;
    setBusy(true);
    setError("");
    try {
      const imported = await importLeagueRules(file);
      setImportReview({ result: imported, beforeRules: cloneLeagueRules(rules) });
      setRules((current) => applyLeagueRuleImport(current, imported));
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "The league rules could not be imported.");
    } finally {
      setBusy(false);
    }
  }

  function undoImport() {
    if (!importReview) return;
    setRules(cloneLeagueRules(importReview.beforeRules));
    setImportReview(undefined);
  }

  return (
    <section className="league-rules-import" aria-labelledby="league-rules-import-title">
      <h3 id="league-rules-import-title">Import league rules</h3>
      <p>
        Start with a PDF or CSV, including a scanned PDF when local OCR is available, or skip this and enter everything
        yourself. DraftMeld fills only settings it recognizes and keeps your draft position and team names manual.
      </p>
      <form className="rule-import" onSubmit={handleImport}>
        <fieldset disabled={busy}>
          <legend>Choose a league rules file</legend>
          <p className="field-help">
            The file is discarded after extraction. Review every imported value before creating the league.
          </p>
          <FormField
            label="League rules PDF or CSV"
            help="Maximum file size: 20 MiB. Scanned PDFs use local OCR when it is available."
          >
            <input
              type="file"
              accept=".pdf,.csv,application/pdf,text/csv"
              onChange={(event) => setFile(event.target.files?.[0] ?? null)}
            />
          </FormField>
          <Button type="submit" disabled={!file || busy}>
            {busy ? "Reading rules and scanned pages..." : "Import and apply recognized settings"}
          </Button>
        </fieldset>
      </form>

      {error ? <StatusMessage tone="error">{error}</StatusMessage> : null}
      {importReview ? <ImportResults result={importReview.result} onUndo={undoImport} /> : null}
    </section>
  );
}

function ImportResults({ result, onUndo }: { result: LeagueRuleImport; onUndo: () => void }) {
  const total = result.settingMatches.length + result.matches.length;
  return (
    <section className="rule-import-results" aria-labelledby="imported-rules-title">
      <h4 id="imported-rules-title">Imported values to review</h4>
      <StatusMessage tone="success">
        Applied {total} recognized {total === 1 ? "value" : "values"}: {result.settingMatches.length} league settings
        and {result.matches.length} scoring values.
      </StatusMessage>
      {result.settingMatches.length ? (
        <ImportTable
          label="Imported league settings"
          caption={`League settings recognized from the ${result.fileType.toUpperCase()} file`}
          rows={result.settingMatches}
          valueHeading="Imported value"
        />
      ) : null}
      {result.matches.length ? (
        <ImportTable
          label="Imported scoring values"
          caption={`Scoring rules recognized from the ${result.fileType.toUpperCase()} file`}
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
            <th scope="col">Detected text</th>
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
