import { useEffect, useRef, useState, type Dispatch, type RefObject, type SetStateAction } from "react";
import { importESPNLeagueRules, importLeagueRules } from "../shared/api/leagues";
import type { LeagueRuleImport, LeagueRules } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { FormField } from "../shared/ui/FormField";
import { StatusMessage } from "../shared/ui/StatusMessage";
import { applyLeagueRuleImport } from "./applyLeagueRuleImport";
import { cloneLeagueRules } from "./leagueDefaults";
import type { OnboardingImportReview } from "./onboarding";

interface LeagueRulesImportProps {
  headingLevel?: 2 | 3;
  rules: LeagueRules;
  setRules: Dispatch<SetStateAction<LeagueRules>>;
  importReview?: OnboardingImportReview;
  setImportReview: Dispatch<SetStateAction<OnboardingImportReview | undefined>>;
}

type ImportMethod = "espn-url" | "espn-paste" | "espn-file" | "file";

const importMethods: Array<{ value: ImportMethod; label: string; description: string }> = [
  { value: "espn-url", label: "ESPN league URL", description: "For leagues ESPN makes publicly readable." },
  { value: "espn-paste", label: "Paste ESPN settings", description: "For private leagues you can open in ESPN." },
  { value: "espn-file", label: "ESPN JSON or PDF", description: "Use an ESPN export or print-to-PDF file." },
  { value: "file", label: "Other PDF or CSV", description: "Use a file from another league provider." },
];

export function LeagueRulesImport({
  headingLevel = 3,
  rules,
  setRules,
  importReview,
  setImportReview,
}: LeagueRulesImportProps) {
  const [method, setMethod] = useState<ImportMethod>();
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

  async function handleImport() {
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
    if (!method) throw new Error("Choose an import method.");
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

  function chooseMethod(nextMethod: ImportMethod) {
    if (method === nextMethod) return;
    setMethod(nextMethod);
    setError("");
    setFile(null);
  }

  const canSubmit =
    method === "espn-url"
      ? leagueURL.trim().length > 0
      : method === "espn-paste"
        ? pastedSettings.trim().length > 0
        : file !== null;
  const Heading = headingLevel === 2 ? "h2" : "h3";

  return (
    <section className="league-rules-import" aria-labelledby="league-rules-import-title">
      <Heading id="league-rules-import-title">Import league settings</Heading>
      <p>
        DraftMeld applies only recognized settings. You review the results before saving, and you can undo the entire
        import.
      </p>
      <div className="rule-import">
        <fieldset disabled={busy}>
          <legend>Choose how to import</legend>
          <p className="field-help">Choose one method. DraftMeld will then show only the fields you need.</p>
          <div className="import-method-options">
            {importMethods.map((option) => (
              <Button
                className="import-method-option"
                variant="neutral"
                key={option.value}
                aria-pressed={method === option.value}
                onClick={() => chooseMethod(option.value)}
                onKeyDown={(event) => {
                  if (event.key === "Enter" || event.key === " ") {
                    event.preventDefault();
                    chooseMethod(option.value);
                  }
                }}
              >
                <span>
                  <strong>{option.label}</strong>
                  <small>{option.description}</small>
                </span>
              </Button>
            ))}
          </div>

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

          <Button variant="primary" disabled={!canSubmit || busy} onClick={() => void handleImport()}>
            {busy ? "Reading league settings…" : "Import and review"}
          </Button>
        </fieldset>
      </div>

      {error ? (
        <StatusMessage tone="error">
          {error} {method === "espn-url" ? "Private leagues can use the paste or ESPN PDF method instead." : ""}
        </StatusMessage>
      ) : null}
      {importReview ? (
        <ImportResults
          headingLevel={headingLevel === 2 ? 3 : 4}
          headingRef={resultsHeading}
          result={importReview.result}
          onUndo={undoImport}
        />
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
  headingLevel,
  headingRef,
  result,
  onUndo,
}: {
  headingLevel: 3 | 4;
  headingRef: RefObject<HTMLHeadingElement | null>;
  result: LeagueRuleImport;
  onUndo: () => void;
}) {
  const total = result.settingMatches.length + result.matches.length;
  const sourceName = formatImportSource(result.fileType);
  const Heading = headingLevel === 3 ? "h3" : "h4";
  return (
    <section className="rule-import-results" aria-labelledby="imported-rules-title">
      <Heading id="imported-rules-title" ref={headingRef} tabIndex={-1}>
        Imported values to review
      </Heading>
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
