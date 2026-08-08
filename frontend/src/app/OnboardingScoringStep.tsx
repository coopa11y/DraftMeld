import { useState, type Dispatch, type FormEvent, type SetStateAction } from "react";
import { importLeagueRules } from "../shared/api/leagues";
import type { LeagueRuleImport, LeagueRules } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { FormField } from "../shared/ui/FormField";
import { StatusMessage } from "../shared/ui/StatusMessage";
import { ScoringSettings } from "./ScoringSettings";
import { applyReceptionPreset, receptionPreset } from "./scoring";

interface OnboardingScoringStepProps {
  rules: LeagueRules;
  setRules: Dispatch<SetStateAction<LeagueRules>>;
}

export function OnboardingScoringStep({ rules, setRules }: OnboardingScoringStepProps) {
  const [file, setFile] = useState<File | null>(null);
  const [result, setResult] = useState<LeagueRuleImport | null>(null);
  const [beforeImport, setBeforeImport] = useState<LeagueRules["scoringRules"] | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  async function handleImport(event: FormEvent) {
    event.preventDefault();
    if (!file) return;
    setBusy(true);
    setError("");
    try {
      const imported = await importLeagueRules(file);
      setBeforeImport({ ...rules.scoringRules });
      setRules((current) => ({
        ...current,
        scoringRules: { ...current.scoringRules, ...imported.rules },
      }));
      setResult(imported);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "The league rules could not be imported.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <section aria-labelledby="scoring-step-title">
      <h2 id="scoring-step-title">Scoring</h2>
      <p>Choose a familiar reception format, import your league document, or enter values manually.</p>
      <FormField label="Reception scoring preset" help="This changes only reception and TE-premium values.">
        <select
          value={receptionPreset(rules.scoringRules)}
          onChange={(event) =>
            setRules((current) => ({
              ...current,
              scoringRules: applyReceptionPreset(current.scoringRules, event.target.value),
            }))
          }
        >
          <option value="0">Standard</option>
          <option value="0.5">Half PPR</option>
          <option value="1">PPR</option>
          <option value="te-premium">TE premium PPR (+0.5)</option>
          <option value="custom">Custom</option>
        </select>
      </FormField>

      <form className="rule-import" onSubmit={handleImport}>
        <fieldset disabled={busy}>
          <legend>Import scoring from a file</legend>
          <p className="field-help">
            Choose a selectable-text PDF or a CSV. DraftMeld discards the file after extraction and applies only
            recognized scoring values. You review the results before creating the league.
          </p>
          <FormField label="League rules PDF or CSV" help="Maximum file size: 20 MiB. Scanned PDFs require OCR first.">
            <input
              type="file"
              accept=".pdf,.csv,application/pdf,text/csv"
              onChange={(event) => setFile(event.target.files?.[0] ?? null)}
            />
          </FormField>
          <Button type="submit" disabled={!file || busy}>
            {busy ? "Reading rules..." : "Import and apply recognized rules"}
          </Button>
        </fieldset>
      </form>

      {error ? <StatusMessage tone="error">{error}</StatusMessage> : null}
      {result ? (
        <section className="rule-import-results" aria-labelledby="imported-rules-title">
          <h3 id="imported-rules-title">Imported values to review</h3>
          <StatusMessage tone="success">Applied {result.matches.length} recognized scoring values.</StatusMessage>
          <div className="table-scroll" role="region" aria-label="Imported scoring values" tabIndex={0}>
            <table>
              <caption>Rules recognized from the {result.fileType.toUpperCase()} file</caption>
              <thead>
                <tr>
                  <th scope="col">Rule</th>
                  <th scope="col">Points</th>
                  <th scope="col">Confidence</th>
                  <th scope="col">Detected text</th>
                </tr>
              </thead>
              <tbody>
                {result.matches.map((match) => (
                  <tr key={match.key}>
                    <th scope="row">{match.label}</th>
                    <td>{match.value}</td>
                    <td>{match.confidence}</td>
                    <td>{match.source}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {result.warnings.map((warning) => (
            <StatusMessage key={warning}>{warning}</StatusMessage>
          ))}
          {beforeImport ? (
            <Button
              onClick={() => {
                setRules((current) => ({ ...current, scoringRules: beforeImport }));
                setResult(null);
                setBeforeImport(null);
              }}
            >
              Undo imported values
            </Button>
          ) : null}
        </section>
      ) : null}

      <ScoringSettings rules={rules} setRules={setRules} disabled={busy} />
    </section>
  );
}
