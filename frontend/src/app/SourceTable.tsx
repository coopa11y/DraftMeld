import type { FormEvent } from "react";
import type {
  LeagueRules,
  ProjectionSource,
  RankingSource,
  RankingSourcePreference,
  RankingSourceRecommendations,
} from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { ConsensusMethodField } from "./ConsensusMethodField";

const dateFormatter = new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeStyle: "short" });

interface SourceTableProps {
  busy: boolean;
  busySourceId: string;
  consensusMethod: LeagueRules["consensusMethod"];
  enabledSourceCount: number;
  preferences: Record<string, RankingSourcePreference>;
  preferencesChanged: boolean;
  projectionSources: ProjectionSource[];
  recommendations: RankingSourceRecommendations;
  sources: RankingSource[];
  onConsensusMethodChange: (method: LeagueRules["consensusMethod"]) => void;
  onDetails: (sourceId: string, kind: "ranking" | "projection") => void;
  onImport: (kind: "pdf" | "ranking-csv" | "projection-csv") => void;
  onPreferencesChange: (preferences: Record<string, RankingSourcePreference>) => void;
  onRefreshSource: (sourceId: string) => void;
  onRefreshProjectionSource: (sourceId: string) => void;
  onReset: () => void;
  onApplyRecommendations: () => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
}

export function SourceTable(props: SourceTableProps) {
  function updatePreference(source: RankingSource, update: Partial<RankingSourcePreference>) {
    const current = props.preferences[source.id] ?? {
      weight: source.defaultWeight,
      enabled: source.defaultEnabled,
    };
    props.onPreferencesChange({ ...props.preferences, [source.id]: { ...current, ...update } });
  }

  const visibleSources = props.sources.filter((source) => {
    if (!source.variantGroup) return true;
    return props.recommendations.sources.find((item) => item.sourceId === source.id)?.fit !== "not recommended";
  });

  return (
    <form className="sources-form" onSubmit={props.onSubmit}>
      <div className="source-settings-bar">
        <ConsensusMethodField
          disabled={props.busy}
          value={props.consensusMethod}
          onChange={props.onConsensusMethodChange}
        />
        <div>
          <p>
            League profile: <strong>{props.recommendations.profile}</strong>
          </p>
          <p>Equal weights have equal influence. Excluded rankings are still checked for outliers.</p>
        </div>
      </div>
      <div className="table-scroll" role="region" aria-label="Ranking and projection sources" tabIndex={0}>
        <table className="source-table">
          <caption>Sources used by DraftMeld</caption>
          <thead>
            <tr>
              <th scope="col">Source</th>
              <th scope="col">Type</th>
              <th scope="col">League fit</th>
              <th scope="col">Status</th>
              <th scope="col">Players</th>
              <th scope="col">Included</th>
              <th scope="col">Weight</th>
              <th scope="col">Actions</th>
            </tr>
          </thead>
          <tbody>
            {visibleSources.map((source) => {
              const preference = props.preferences[source.id] ?? {
                weight: source.defaultWeight,
                enabled: source.defaultEnabled,
              };
              const updating = props.busySourceId === source.id;
              const recommendation = props.recommendations.sources.find((item) => item.sourceId === source.id);
              return (
                <tr key={source.id}>
                  <th scope="row">{source.name}</th>
                  <td>{rankingType(source)}</td>
                  <td>
                    {recommendation ? (
                      <details>
                        <summary>{recommendation.fit}</summary>
                        <p>{recommendation.reason}</p>
                      </details>
                    ) : (
                      "User source"
                    )}
                  </td>
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
                        aria-label={`Include ${source.name} in consensus`}
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
                          aria-label={`Weight for ${source.name}`}
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
                          aria-label={`Update ${source.name} now`}
                          disabled={props.busy || Boolean(props.busySourceId)}
                          onClick={() => props.onRefreshSource(source.id)}
                        >
                          {updating ? "Updating…" : "Update now"}
                        </Button>
                      ) : (
                        <Button
                          aria-label={`${source.recordCount > 0 ? "Replace" : "Import"} ${source.name}`}
                          disabled={props.busy}
                          onClick={() => props.onImport(source.importMode === "pdf-upload" ? "pdf" : "ranking-csv")}
                        >
                          {source.recordCount > 0 ? "Replace file" : "Import"}
                        </Button>
                      )}
                      <Button
                        aria-label={`Details for ${source.name}`}
                        onClick={() => props.onDetails(source.id, "ranking")}
                      >
                        Details
                      </Button>
                    </div>
                  </td>
                </tr>
              );
            })}
            {props.projectionSources.map((source) => (
              <tr key={source.id}>
                <th scope="row">{source.name}</th>
                <td>{source.importMode === "download" ? "Online projection" : "CSV projection"}</td>
                <td>{source.importMode === "download" ? "League-scored raw statistics" : "Applies league scoring"}</td>
                <td>{source.recordCount > 0 ? formatUpdated(source.importedAt) : "Not imported"}</td>
                <td>{source.recordCount}</td>
                <td>Scoring</td>
                <td>—</td>
                <td>
                  <div className="table-actions">
                    {source.importMode === "download" ? (
                      <Button
                        aria-label={`Update ${source.name} now`}
                        disabled={props.busy || Boolean(props.busySourceId)}
                        onClick={() => props.onRefreshProjectionSource(source.id)}
                      >
                        {props.busySourceId === source.id ? "Updatingâ€¦" : "Update now"}
                      </Button>
                    ) : (
                      <Button
                        aria-label={`Replace ${source.name}`}
                        disabled={props.busy}
                        onClick={() => props.onImport("projection-csv")}
                      >
                        Replace file
                      </Button>
                    )}
                    <Button
                      aria-label={`Details for ${source.name}`}
                      onClick={() => props.onDetails(source.id, "projection")}
                    >
                      Details
                    </Button>
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
        <Button disabled={props.busy} onClick={props.onApplyRecommendations}>
          Apply league recommendations
        </Button>
        <Button disabled={props.busy || !props.preferencesChanged} onClick={props.onReset}>
          Discard changes
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
