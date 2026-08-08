import { useMemo } from "react";
import type { DraftSnapshot, Pick } from "../shared/api/types";
import { Panel } from "../shared/ui/Panel";

interface DraftOverviewProps {
  snapshot: DraftSnapshot;
}

export function DraftOverview({ snapshot }: DraftOverviewProps) {
  return (
    <section className="draft-overview" aria-labelledby="draft-overview-title">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Every pick and roster</p>
          <h2 id="draft-overview-title">League draft room</h2>
        </div>
        <span>
          {snapshot.isComplete ? "Draft complete" : `${snapshot.history.length} of ${snapshot.totalPicks} picks`}
        </span>
      </div>

      <Panel variant="board" aria-labelledby="draft-grid-title">
        <h3 id="draft-grid-title">{snapshot.draftType === "auction" ? "Draft ledger" : "Draft grid"}</h3>
        {snapshot.draftType === "auction" ? <AuctionLedger snapshot={snapshot} /> : <DraftGrid snapshot={snapshot} />}
      </Panel>

      <div className="team-roster-grid" aria-label="League team rosters">
        {snapshot.teams.map((team) => (
          <Panel key={team.number} variant="side" aria-labelledby={`team-${team.number}-title`}>
            <p className="eyebrow">
              {team.roster.length} players{team.isUser ? " · Your team" : ""}
            </p>
            <h3 id={`team-${team.number}-title`}>{team.name} roster</h3>
            {team.roster.length ? (
              <ol className="team-list">
                {team.roster.map((player) => (
                  <li key={player.id}>
                    <strong>{player.position}</strong>
                    <span>
                      {player.name}, {player.nflTeam}
                    </span>
                  </li>
                ))}
              </ol>
            ) : (
              <p className="empty-state">No players drafted.</p>
            )}
          </Panel>
        ))}
      </div>
    </section>
  );
}

function DraftGrid({ snapshot }: DraftOverviewProps) {
  const rounds = Math.ceil(snapshot.totalPicks / snapshot.teams.length);
  const picksByRoundAndTeam = useMemo(() => {
    const indexed = new Map<string, Pick>();
    for (const pick of snapshot.history) {
      const round = Math.floor((pick.number - 1) / snapshot.teams.length) + 1;
      indexed.set(`${round}:${pick.teamNumber}`, pick);
    }
    return indexed;
  }, [snapshot.history, snapshot.teams.length]);
  return (
    <div className="table-scroll" role="region" aria-label="Full draft grid" tabIndex={0}>
      <table className="draft-grid">
        <caption>All draft picks organized by round and team</caption>
        <thead>
          <tr>
            <th scope="col">Round</th>
            {snapshot.teams.map((team) => (
              <th scope="col" key={team.number}>
                {team.name}
                {team.isUser ? " (you)" : ""}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {Array.from({ length: rounds }, (_, index) => index + 1).map((round) => (
            <tr key={round}>
              <th scope="row">{round}</th>
              {snapshot.teams.map((team) => (
                <DraftGridCell key={team.number} pick={picksByRoundAndTeam.get(`${round}:${team.number}`)} />
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function DraftGridCell({ pick }: { pick?: Pick }) {
  return (
    <td>
      {pick ? (
        <>
          <strong>{pick.player.name}</strong>
          <span>
            {pick.player.position} · Pick {pick.number}
          </span>
        </>
      ) : (
        <span className="empty-pick">—</span>
      )}
    </td>
  );
}

function AuctionLedger({ snapshot }: DraftOverviewProps) {
  return (
    <div className="table-scroll" role="region" aria-label="Full auction draft ledger" tabIndex={0}>
      <table className="draft-grid">
        <caption>Every auction result in recorded order</caption>
        <thead>
          <tr>
            <th scope="col">Pick</th>
            <th scope="col">Team</th>
            <th scope="col">Player</th>
            <th scope="col">Position</th>
            <th scope="col">Cost</th>
          </tr>
        </thead>
        <tbody>
          {snapshot.history.map((pick) => (
            <tr key={pick.eventId}>
              <th scope="row">{pick.number}</th>
              <td>{pick.teamName}</td>
              <td>{pick.player.name}</td>
              <td>{pick.player.position}</td>
              <td>${pick.cost.toFixed(0)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
