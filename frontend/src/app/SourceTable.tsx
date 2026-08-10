import type { FormEvent } from "react";
import type { LeagueRules, ProjectionSource, RankingSource, RankingSourcePreference } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { FormField } from "../shared/ui/FormField";

const dateFormatter = new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeStyle: "short" });

interface SourceTableProps {
  busy: boolean;
  busySourceId: string;
  consensusMethod: LeagueRules["consensusMethod"];
  enabledSourceCount: number;
  preferences: Record<string, RankingSourcePreference>;
  preferencesChanged: boolean;
  projectionSources: ProjectionSource[];
  sources: RankingSource[];
  onConsensusMethodChange: (method: LeagueRules["consensusMethod"]) => void;
  onDetails: (sourceId: string, kind: "ranking" | "projection") => void;
  onImport: (kind: "pdf" | "ranking-csv" | "projection-csv") => void;
  onPreferencesChange: (preferences: Record<string, RankingSourcePreference>) => void;
  onRefreshSource: (sourceId: string) => void;
  onReset: () => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
}

export function SourceTable(props: SourceTableProps) {
  function updatePreference(source: RankingSource, update: Partial<RankingSourcePreference>) {
    const current = props.preferences[source.id] ?? { weight: source.defaultWeight, enabled: true };
    props.onPreferencesChange({ ...props.preferences, [source.id]: { ...current, ...update } });
  }

  return (
    <form className="sources-form" onSubmit={props.onSubmit}>
      <div className="source-settings-bar">
        <FormField label="Consensus method">
          <select
            value={props.consensusMethod}
            disabled={props.busy}
            onChange={(event) => props.onConsensusMethodChange(event.target.value as LeagueRules["consensusMethod"])}
          >
            <option value="weighted-median">Weighted median</option>
            <option value="trimmed-mean">Trimmed mean</option>
            <option value="weighted-average">Weighted average</option>
          </select>
        </FormField>
        <p>Equal weights have equal influence. Excluded rankings are still checked for outliers.</p>
      </div>
      <div className="table-scroll" role="region" aria-label="Ranking and projection sources" tabIndex={0}>
        <table className="source-table">
          <caption>Sources used by DraftMeld</caption>
          <thead>
            <tr>
              <th scope="col">Source</th>
              <th scope="col">Type</th>
              <th scope="col">Status</th>
              <th scope="col">Players</th>
              <th scope="col">Included</th>
              <th scope="col">Weight</th>
              <th scope="col">Actions</th>
            </tr>
          </thead>
          <tbody>
            {props.sources.map((source) => {
              const preference = props.preferences[source.id] ?? { weight: source.defaultWeight, enabled: true };
              const updating = props.busySourceId === source.id;
              return (
                <tr key={source.id}>
                  <th scope="row">{source.name}</th>
                  <td>{rankingType(source)}</td>
                  <td>
                    {updating
                      ? "Updating"
                      : source.recordCount > 0
                        ? formatUpdated(source.refreshedAt)
                        : "Not imported"}
                  </td>
                  <td>{source.recordCount}</td>
                  <td>
                    <label className="compact-check">
                      <input
                        type="checkbox"
                        aria-label="Include in consensus"
                        checked={preference.enabled}
                        disabled={props.busy || updating || (preference.enabled && props.enabledSourceCount === 1)}
                        onChange={(event) => updatePreference(source, { enabled: event.target.checked })}
                      />
                      <span aria-hidden="true">{preference.enabled ? "Yes" : "No"}</span>
                    </label>
                  </td>
                  <td>
                    {preference.enabled ? (
                      <label className="compact-weight">
                        <input
                          type="number"
                          aria-label="Weight"
                          min="0.1"
                          max="10"
                          step="0.1"
                          required
                          value={preference.weight}
                          disabled={props.busy || updating}
                          onChange={(event) => updatePreference(source, { weight: Number(event.target.value) })}
                        />
                      </label>
                    ) : (
                      <span className="muted-value">Not used</span>
                    )}
                  </td>
                  <td>
                    <div className="table-actions">
                      {source.importMode === "download" ? (
                        <Button
                          disabled={props.busy || Boolean(props.busySourceId)}
                          onClick={() => props.onRefreshSource(source.id)}
                        >
                          {updating ? "Updating…" : "Update now"}
                        </Button>
                      ) : (
                        <Button
                          disabled={props.busy}
                          onClick={() => props.onImport(source.importMode === "pdf-upload" ? "pdf" : "ranking-csv")}
                        >
                          {source.recordCount > 0 ? "Replace file" : "Import"}
                        </Button>
                      )}
                      <Button onClick={() => props.onDetails(source.id, "ranking")}>Details</Button>
                    </div>
                  </td>
                </tr>
              );
            })}
            {props.projectionSources.map((source) => (
              <tr key={source.id}>
                <th scope="row">{source.name}</th>
                <td>CSV projection</td>
                <td>{formatUpdated(source.importedAt)}</td>
                <td>{source.recordCount}</td>
                <td>Scoring</td>
                <td>—</td>
                <td>
                  <div className="table-actions">
                    <Button disabled={props.busy} onClick={() => props.onImport("projection-csv")}>
                      Replace file
                    </Button>
                    <Button onClick={() => props.onDetails(source.id, "projection")}>Details</Button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div className="source-save-bar">
        <Button variant="primary" type="submit" disabled={props.busy || !props.preferencesChanged}>
          Save preferences
        </Button>
        <Button disabled={props.busy} onClick={props.onReset}>
          Restore defaults
        </Button>
      </div>
    </form>
  );
}

function rankingType(source: RankingSource) {
  if (source.importMode === "pdf-upload") return "PDF ranking";
  if (source.importMode === "csv-upload") return "CSV ranking";
  return source.role === "ranking" ? "Online ranking" : `Online ${source.role}`;
}

function formatUpdated(value?: string | null) {
  return value ? dateFormatter.format(new Date(value)) : "Not updated";
}
