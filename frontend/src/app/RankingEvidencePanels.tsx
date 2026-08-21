import type { ConsensusRanking, WatchlistPlayer } from "../shared/api/types";
import { Panel } from "../shared/ui/Panel";

const providerCoverage = [
  { provider: "CBS Sports", status: "Connected", detail: "Current public PPR Top 200 consensus." },
  {
    provider: "Yahoo Fantasy",
    status: "Public Standard",
    detail: "The public Standard Top 200 is available; league-specific XRank requires Yahoo authorization.",
  },
  {
    provider: "NFL.com",
    status: "Held out",
    detail: "The public overall draft table still contains the prior-season board.",
  },
  {
    provider: "ESPN",
    status: "Online PPR",
    detail: "The current public PPR player service is connected, with PDF import retained as a fallback.",
  },
  {
    provider: "Draft Sharks",
    status: "League-matched presets",
    detail:
      "Public scoring and quarterback presets include tiers and projection evidence; private league sync requires the user's authorized export.",
  },
] as const;

interface RankingEvidencePanelsProps {
  enabledImportedSourceCount: number;
  leagueName: string;
  rankings: ConsensusRanking[];
  watchlist: WatchlistPlayer[];
}

export function RankingEvidencePanels({
  enabledImportedSourceCount,
  leagueName,
  rankings,
  watchlist,
}: RankingEvidencePanelsProps) {
  return (
    <>
      {watchlist.length > 0 ? <WatchlistPanel players={watchlist} /> : null}
      <ProviderCoveragePanel />
      {rankings.length > 0 ? (
        <ConsensusPreview
          enabledImportedSourceCount={enabledImportedSourceCount}
          leagueName={leagueName}
          rankings={rankings}
        />
      ) : null}
    </>
  );
}

function WatchlistPanel({ players }: { players: WatchlistPlayer[] }) {
  return (
    <Panel variant="ranking" className="watchlist-panel" aria-labelledby="watchlist-heading">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Excluded-source signals</p>
          <h2 id="watchlist-heading">Worth another look</h2>
          <p className="section-description">
            A short list of players ranked meaningfully higher by sources you excluded from the main consensus.
          </p>
        </div>
      </div>
      <ul className="watchlist-grid">
        {players.map((player) => (
          <li key={player.playerKey}>
            <article>
              <h3>
                {player.name}{" "}
                <span>
                  {player.position} - {player.team}
                </span>
              </h3>
              <ul>
                {player.signals.map((signal) => (
                  <li key={signal.sourceId}>
                    {player.consensusRank === null
                      ? `${signal.sourceName} ranks this player #${signal.sourceRank}; enabled sources do not currently rank the player.`
                      : `${signal.sourceName} ranks this player #${signal.sourceRank}, ${signal.spotsHigher} spots above consensus #${player.consensusRank}.`}
                  </li>
                ))}
              </ul>
            </article>
          </li>
        ))}
      </ul>
    </Panel>
  );
}

function ProviderCoveragePanel() {
  return (
    <Panel variant="ranking" aria-labelledby="provider-coverage-heading">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Freshness safeguards</p>
          <h2 id="provider-coverage-heading">Platform connector status</h2>
          <p className="section-description">
            DraftMeld holds a provider out when its available data is stale, incomplete, or requires approval.
          </p>
        </div>
      </div>
      <ul className="coverage-list">
        {providerCoverage.map((provider) => (
          <li key={provider.provider}>
            <div>
              <strong>{provider.provider}</strong>
              <span>{provider.status}</span>
            </div>
            <p>{provider.detail}</p>
          </li>
        ))}
      </ul>
    </Panel>
  );
}

function ConsensusPreview({
  enabledImportedSourceCount,
  leagueName,
  rankings,
}: Omit<RankingEvidencePanelsProps, "watchlist">) {
  return (
    <Panel variant="ranking" aria-labelledby="consensus-heading">
      <div className="section-heading">
        <div>
          <p className="eyebrow">League-weighted preview</p>
          <h2 id="consensus-heading">DraftMeld consensus for {leagueName}</h2>
          <p className="section-description">
            Lists are normalized for source depth, primary ranking omissions are handled conservatively, and contextual
            market or usage signals only apply when present. Lower is better.
          </p>
        </div>
      </div>
      <div className="table-scroll" role="region" aria-label="Consensus ranking preview" tabIndex={0}>
        <table>
          <caption>Top 25 blended player rankings</caption>
          <thead>
            <tr>
              <th scope="col">Rank</th>
              <th scope="col">Player</th>
              <th scope="col">Pos.</th>
              <th scope="col">Sources</th>
              <th scope="col">Score</th>
              <th scope="col">Confidence</th>
            </tr>
          </thead>
          <tbody>
            {rankings.slice(0, 25).map((player) => (
              <tr key={player.playerKey}>
                <td>{player.rank}</td>
                <th scope="row">
                  <span className="player-name">{player.name}</span>
                  <span className="player-meta">{player.team || "Team unavailable"}</span>
                  {player.projection ? <ProjectionEvidenceDetails evidence={player.projection} /> : null}
                </th>
                <td>{player.position}</td>
                <td>
                  {player.sourceCount} of {enabledImportedSourceCount}
                </td>
                <td>{player.score.toFixed(1)}</td>
                <td>
                  <span className={`confidence-badge confidence-${player.confidence}`}>{player.confidence}</span>
                  <span className="player-meta">
                    {Math.round(player.coverage * 100)}% coverage, {player.rankRange}-rank range
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </Panel>
  );
}

function ProjectionEvidenceDetails({ evidence }: { evidence: NonNullable<ConsensusRanking["projection"]> }) {
  return (
    <details className="projection-evidence-details">
      <summary>{evidence.sourceName} projection details</summary>
      <dl>
        <div>
          <dt>Profile</dt>
          <dd>{evidence.profile}</dd>
        </div>
        <div>
          <dt>Projection range</dt>
          <dd>
            {evidence.floorProjection.toFixed(1)} floor, {evidence.sourceProjection.toFixed(1)} Draft Sharks,{" "}
            {evidence.ceilingProjection.toFixed(1)} ceiling
          </dd>
        </div>
        <div>
          <dt>Consensus projection</dt>
          <dd>{evidence.consensusProjection.toFixed(1)}</dd>
        </div>
        <div>
          <dt>3D Value</dt>
          <dd>{evidence.sourceValue.toFixed(1)}</dd>
        </div>
        <div>
          <dt>Availability</dt>
          <dd>
            {evidence.games} games, bye week {evidence.byeWeek || "unavailable"}
          </dd>
        </div>
        <div>
          <dt>Risk and schedule</dt>
          <dd>
            {evidence.injuryRisk.toFixed(1)}% injury risk, {signedPercent(evidence.scheduleStrength)} schedule strength
          </dd>
        </div>
      </dl>
    </details>
  );
}

function signedPercent(value: number) {
  return `${value > 0 ? "+" : ""}${value.toFixed(1)}%`;
}
