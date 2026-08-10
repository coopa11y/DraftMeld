import type { League } from "../shared/api/types";
import { useViewHeadingFocus } from "../shared/hooks/useViewHeadingFocus";
import { Button } from "../shared/ui/Button";
import { Panel } from "../shared/ui/Panel";

interface LeagueOverviewProps {
  busy: boolean;
  league: League;
  onBack: () => void;
  onEdit: () => void;
  onMockDraft: () => void;
  onOpenDraft: () => void;
  onOpenSources: () => void;
}

export function LeagueOverview(props: LeagueOverviewProps) {
  const heading = useViewHeadingFocus<HTMLHeadingElement>();
  const rosterSpots = props.league.rosterSlots.reduce((total, slot) => total + slot.count, 0);
  const namedTeams = props.league.teamNames.filter((name) => name.trim()).length;
  const includedSources = Object.values(props.league.sourcePreferences).filter((source) => source.enabled).length;
  const needsDraftPosition = props.league.draftType !== "auction" && props.league.draftPosition === 0;

  return (
    <Panel variant="form" className="league-overview" aria-labelledby="league-overview-title">
      <Button onClick={props.onBack}>Back to all leagues</Button>
      <div className="form-heading">
        <div>
          <p className="eyebrow">League home</p>
          <h1 id="league-overview-title" ref={heading} tabIndex={-1}>
            {props.league.name}
          </h1>
          <p>Review setup, adjust league rules, or deliberately enter a real or mock draft.</p>
        </div>
        <Button variant="primary" onClick={props.onEdit}>
          Edit league setup
        </Button>
      </div>

      <dl className="league-setup-summary">
        <div>
          <dt>League</dt>
          <dd>
            {props.league.teamCount} teams · {formatLeagueFormat(props.league.leagueFormat)}
          </dd>
        </div>
        <div>
          <dt>Draft</dt>
          <dd>
            {formatDraftType(props.league.draftType)} ·{" "}
            {props.league.draftPosition > 0 ? `Position ${props.league.draftPosition}` : "Position not set"}
          </dd>
        </div>
        <div>
          <dt>Teams</dt>
          <dd>
            {namedTeams} of {props.league.teamCount} named
          </dd>
        </div>
        <div>
          <dt>Roster</dt>
          <dd>{rosterSpots} total spots</dd>
        </div>
        <div>
          <dt>Scoring</dt>
          <dd>{formatScoring(props.league.scoringRules.reception)}</dd>
        </div>
        <div>
          <dt>Sources</dt>
          <dd>{includedSources ? `${includedSources} included` : "Using defaults"}</dd>
        </div>
      </dl>

      <div className="league-overview-actions">
        <section aria-labelledby="setup-actions-title">
          <h2 id="setup-actions-title">Finish setup</h2>
          <p>Configure rules, teams, positions, scoring, and ranking sources before draft day.</p>
          <Button onClick={props.onEdit}>League rules and scoring</Button>
          <Button onClick={props.onOpenSources}>Ranking sources</Button>
        </section>
        <section aria-labelledby="draft-actions-title">
          <h2 id="draft-actions-title">Draft</h2>
          <p>Open your real board, or practice safely in a separate session using this league&apos;s settings.</p>
          <Button variant="primary" onClick={props.onOpenDraft}>
            Open draft board
          </Button>
          <Button disabled={props.busy || needsDraftPosition} onClick={props.onMockDraft}>
            Start mock draft
          </Button>
          {needsDraftPosition ? (
            <p className="field-help">Set your draft position before starting a real or mock draft.</p>
          ) : null}
        </section>
      </div>
      <p className="field-help">
        Mock drafts use a separate session. This league and its draft history stay unchanged.
      </p>
    </Panel>
  );
}

export function formatDraftType(type: League["draftType"]): string {
  return `${type.charAt(0).toUpperCase()}${type.slice(1)} draft`;
}

export function formatScoring(receptions: number): string {
  if (receptions === 1) return "PPR scoring";
  if (receptions === 0.5) return "Half-PPR scoring";
  if (receptions === 0) return "Standard scoring";
  return `${receptions} points per reception`;
}

const formatLeagueFormat = (format: League["leagueFormat"]) => (format === "dynasty" ? "Dynasty" : "Redraft");
