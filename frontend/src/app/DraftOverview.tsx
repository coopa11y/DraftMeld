import { useMemo } from "react";
import type { DraftPickSlot, DraftPickTrade, DraftSnapshot, Pick as DraftPick } from "../shared/api/types";
import { Panel } from "../shared/ui/Panel";
import { DraftPickTrades } from "./DraftPickTrades";

interface DraftOverviewProps {
  snapshot: DraftSnapshot;
  busy: boolean;
  onCreateTrade: (
    teamOne: number,
    teamTwo: number,
    teamOneReceives: number[],
    teamTwoReceives: number[],
  ) => Promise<void>;
  onDeleteTrade: (trade: DraftPickTrade) => Promise<void>;
}

export function DraftOverview({ snapshot, busy, onCreateTrade, onDeleteTrade }: DraftOverviewProps) {
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

      {snapshot.draftType === "auction" ? null : (
        <DraftPickTrades snapshot={snapshot} busy={busy} onCreate={onCreateTrade} onDelete={onDeleteTrade} />
      )}

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

function DraftGrid({ snapshot }: Pick<DraftOverviewProps, "snapshot">) {
  const rounds = Math.ceil(snapshot.totalPicks / snapshot.teams.length);
  const picksByNumber = useMemo(() => {
    const indexed = new Map<number, DraftPick>();
    for (const pick of snapshot.history) {
      indexed.set(pick.number, pick);
    }
    return indexed;
  }, [snapshot.history]);
  const slotsByRoundAndOriginalTeam = useMemo(
    () => new Map(snapshot.pickSlots.map((slot) => [`${slot.round}:${slot.originalTeamNumber}`, slot])),
    [snapshot.pickSlots],
  );
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
              {snapshot.teams.map((team) => {
                const slot = slotsByRoundAndOriginalTeam.get(`${round}:${team.number}`);
                return (
                  <DraftGridCell
                    key={team.number}
                    slot={slot}
                    pick={slot ? picksByNumber.get(slot.overallNumber) : undefined}
                  />
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function DraftGridCell({ pick, slot }: { pick?: DraftPick; slot?: DraftPickSlot }) {
  return (
    <td>
      {pick ? (
        <>
          <strong>{pick.player.name}</strong>
          <span>
            {pick.player.position} · Pick {pick.number}
          </span>
          {slot && slot.ownerTeamNumber !== slot.originalTeamNumber ? (
            <span>Traded to {slot.ownerTeamName}</span>
          ) : null}
        </>
      ) : (
        <>
          <span className="empty-pick">{slot ? `Pick ${slot.overallNumber}` : "—"}</span>
          {slot && slot.ownerTeamNumber !== slot.originalTeamNumber ? (
            <span>Traded to {slot.ownerTeamName}</span>
          ) : null}
        </>
      )}
    </td>
  );
}

function AuctionLedger({ snapshot }: Pick<DraftOverviewProps, "snapshot">) {
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
