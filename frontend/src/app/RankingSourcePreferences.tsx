import type { FormEvent } from "react";
import type { LeagueRules, RankingSource, RankingSourcePreference } from "../shared/api/types";

const refreshTimeFormatter = new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeStyle: "short" });

interface RankingSourcePreferencesProps {
  busy: boolean;
  consensusMethod: LeagueRules["consensusMethod"];
  enabledSourceCount: number;
  leagueName: string;
  preferences: Record<string, RankingSourcePreference>;
  preferencesChanged: boolean;
  sources: RankingSource[];
  onConsensusMethodChange: (method: LeagueRules["consensusMethod"]) => void;
  onPreferencesChange: (preferences: Record<string, RankingSourcePreference>) => void;
  onReset: () => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
}

export function RankingSourcePreferences({
  busy,
  consensusMethod,
  enabledSourceCount,
  leagueName,
  preferences,
  preferencesChanged,
  sources,
  onConsensusMethodChange,
  onPreferencesChange,
  onReset,
  onSubmit,
}: RankingSourcePreferencesProps) {
  function updatePreference(sourceId: string, update: Partial<RankingSourcePreference>) {
    const source = sources.find((candidate) => candidate.id === sourceId);
    if (!source) return;
    const current = preferences[sourceId] ?? { weight: source.defaultWeight, enabled: true };
    onPreferencesChange({ ...preferences, [sourceId]: { ...current, ...update } });
  }

  return (
    <form className="source-weight-form" onSubmit={onSubmit}>
      <fieldset disabled={busy}>
        <legend>Source preferences for {leagueName}</legend>
        <p className="field-help">Included sources shape the consensus. Excluded sources stay available for the compact &quot;Worth another look&quot; list. Weight 2 has twice the pull of weight 1, and matching weights have equal influence. At least one source must remain included.</p>
        <label className="consensus-method-control">
          Consensus method
          <select value={consensusMethod} onChange={(event) => onConsensusMethodChange(event.target.value as LeagueRules["consensusMethod"])}>
            <option value="weighted-median">Weighted median — resistant to outliers</option>
            <option value="trimmed-mean">Trimmed mean — ignores extremes</option>
            <option value="weighted-average">Weighted average — maximum source sensitivity</option>
          </select>
        </label>
        <ul className="source-grid">
          {sources.map((source) => {
            const preference = preferences[source.id] ?? { weight: source.defaultWeight, enabled: true };
            return (
              <li key={source.id}>
                <article className={`source-card${preference.enabled ? "" : " source-card-disabled"}`}>
                  <div className="source-card-heading"><h2>{source.name}</h2><span>{source.recordCount > 0 ? `${source.recordCount} players` : "Not imported"}</span></div>
                  <p>{source.description}</p>
                  <dl>
                    <div><dt>Method</dt><dd>{source.methodology}</dd></div>
                    <div><dt>Signal role</dt><dd>{source.role}</dd></div>
                    <div><dt>License</dt><dd>{source.license}</dd></div>
                    <div><dt>Default weight</dt><dd>{source.defaultWeight}</dd></div>
                    <div><dt>Published</dt><dd>{source.publishedAt || "Refresh required"}</dd></div>
                    <div><dt>Last refreshed</dt><dd>{formatRefreshTime(source.refreshedAt)}</dd></div>
                  </dl>
                  <label className="source-enabled-control">
                    <input
                      type="checkbox"
                      checked={preference.enabled}
                      disabled={preference.enabled && enabledSourceCount === 1}
                      onChange={(event) => updatePreference(source.id, { enabled: event.target.checked })}
                    />
                    Include {source.name} in consensus
                  </label>
                  <label className="source-weight-control" htmlFor={`source-weight-${source.id}`}>
                    <span>{source.name} influence</span>
                    <input
                      id={`source-weight-${source.id}`}
                      type="number"
                      min="0.1"
                      max="10"
                      step="0.1"
                      required
                      disabled={!preference.enabled}
                      value={preference.weight}
                      onChange={(event) => updatePreference(source.id, { weight: Number(event.target.value) })}
                    />
                  </label>
                  <a href={source.projectUrl} target="_blank" rel="noreferrer">View source website<span className="sr-only"> for {source.name} (opens in a new tab)</span></a>
                </article>
              </li>
            );
          })}
        </ul>
        <div className="source-weight-actions">
          <button className="primary-button" type="submit" disabled={!preferencesChanged || busy}>Save preferences</button>
          <button className="secondary-button" type="button" onClick={onReset} disabled={busy}>Include all and restore defaults</button>
        </div>
      </fieldset>
    </form>
  );
}

function formatRefreshTime(value?: string | null) {
  return value ? refreshTimeFormatter.format(new Date(value)) : "Refresh required";
}
