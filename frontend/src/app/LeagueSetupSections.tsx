import type { Dispatch, SetStateAction } from "react";
import type { LeagueRules } from "../shared/api/types";
import { FormField } from "../shared/ui/FormField";
import { applyReceptionPreset, receptionPreset } from "./scoring";

type SetRules = Dispatch<SetStateAction<LeagueRules>>;

function resizeTeamNames(names: string[], teamCount: number): string[] {
  return Array.from({ length: teamCount }, (_, index) => names[index] ?? "");
}

function resizeDraftOrder(order: number[], teamCount: number): number[] {
  const retained = order.filter((team) => team <= teamCount);
  for (let team = 1; team <= teamCount; team += 1) {
    if (!retained.includes(team)) retained.push(team);
  }
  return retained;
}

function moveTeamToPosition(order: number[], team: number, position: number): number[] {
  const updated = [...order];
  const currentIndex = updated.indexOf(team);
  const targetIndex = position - 1;
  [updated[currentIndex], updated[targetIndex]] = [updated[targetIndex], updated[currentIndex]];
  return updated;
}

interface SettingsProps {
  rules: LeagueRules;
  setRules: SetRules;
  disabled: boolean;
}

export function LeagueSettings({
  rules,
  setRules,
  disabled,
  lockSeason = false,
}: SettingsProps & { lockSeason?: boolean }) {
  return (
    <fieldset disabled={disabled}>
      <legend>League settings</legend>
      <div className="form-grid">
        <FormField label="League name">
          <input
            required
            maxLength={80}
            value={rules.name}
            onChange={(event) => setRules({ ...rules, name: event.target.value })}
          />
        </FormField>
        <FormField label="Number of teams">
          <input
            type="number"
            min="2"
            max="32"
            required
            value={rules.teamCount}
            onChange={(event) => {
              const teamCount = Number(event.target.value);
              const userTeamNumber = Math.min(rules.userTeamNumber, teamCount);
              const draftOrder = resizeDraftOrder(rules.draftOrder, teamCount);
              const draftPosition = draftOrder.indexOf(userTeamNumber) + 1;
              setRules({
                ...rules,
                teamCount,
                draftPosition,
                userTeamNumber,
                draftOrder,
                teamNames: resizeTeamNames(rules.teamNames, teamCount),
              });
            }}
          />
        </FormField>
        <FormField label="League format" help="Redraft shows only this season. Dynasty can track future rookie picks.">
          <select
            value={rules.leagueFormat}
            onChange={(event) => {
              const leagueFormat = event.target.value as LeagueRules["leagueFormat"];
              setRules({
                ...rules,
                leagueFormat,
                futurePickSeasons: leagueFormat === "dynasty" ? Math.max(1, rules.futurePickSeasons) : 0,
              });
            }}
          >
            <option value="redraft">Redraft</option>
            <option value="dynasty">Dynasty</option>
          </select>
        </FormField>
        <FormField
          label="Current season"
          help={
            lockSeason
              ? "Use the guided season rollover in the draft workspace to start the next dynasty season."
              : undefined
          }
        >
          <input
            type="number"
            min="2020"
            max="2200"
            required
            readOnly={lockSeason}
            value={rules.season}
            onChange={(event) => setRules({ ...rules, season: Number(event.target.value) })}
          />
        </FormField>
        <FormField
          label="Reception scoring preset"
          help="Choose a common format, then customize values in the Scoring section."
        >
          <select
            value={receptionPreset(rules.scoringRules)}
            onChange={(event) => {
              if (event.target.value === "custom") return;
              setRules({ ...rules, scoringRules: applyReceptionPreset(rules.scoringRules, event.target.value) });
            }}
          >
            <option value="0">Standard</option>
            <option value="0.5">Half PPR</option>
            <option value="1">PPR</option>
            <option value="te-premium">TE premium PPR (+0.5)</option>
            <option value="custom">Custom</option>
          </select>
        </FormField>
        <FormField label="Consensus method">
          <select
            value={rules.consensusMethod}
            onChange={(event) =>
              setRules({ ...rules, consensusMethod: event.target.value as LeagueRules["consensusMethod"] })
            }
          >
            <option value="weighted-median">Weighted median</option>
            <option value="trimmed-mean">Trimmed mean</option>
            <option value="weighted-average">Weighted average</option>
          </select>
        </FormField>
      </div>
    </fieldset>
  );
}

export function DraftSettings({ rules, setRules, disabled }: SettingsProps) {
  return (
    <fieldset disabled={disabled}>
      <legend>Draft settings</legend>
      <div className="form-grid">
        <FormField label="Draft format">
          <select
            value={rules.draftType}
            onChange={(event) => setRules({ ...rules, draftType: event.target.value as LeagueRules["draftType"] })}
          >
            <option value="snake">Snake</option>
            <option value="linear">Linear</option>
            <option value="auction">Auction</option>
          </select>
        </FormField>
        <FormField
          label="Your draft position"
          help="Choose the pick number assigned to you. You do not need to know the other teams' names."
        >
          <select
            required
            value={rules.draftPosition}
            onChange={(event) => {
              const draftPosition = Number(event.target.value);
              const draftOrder = moveTeamToPosition(rules.draftOrder, rules.userTeamNumber, draftPosition);
              setRules({
                ...rules,
                draftPosition,
                draftOrder,
              });
            }}
          >
            {Array.from({ length: rules.teamCount }, (_, index) => (
              <option key={index + 1} value={index + 1}>
                Pick {index + 1}
              </option>
            ))}
          </select>
        </FormField>
        {rules.leagueFormat === "dynasty" ? (
          <>
            <FormField label="Future pick seasons" help="How many upcoming seasons of rookie picks can be traded.">
              <select
                value={rules.futurePickSeasons}
                onChange={(event) => setRules({ ...rules, futurePickSeasons: Number(event.target.value) })}
              >
                {[1, 2, 3, 4, 5].map((count) => (
                  <option key={count} value={count}>
                    {count}
                  </option>
                ))}
              </select>
            </FormField>
            <FormField label="Rookie draft rounds">
              <select
                value={rules.rookieDraftRounds}
                onChange={(event) => setRules({ ...rules, rookieDraftRounds: Number(event.target.value) })}
              >
                {Array.from({ length: 10 }, (_, index) => index + 1).map((round) => (
                  <option key={round} value={round}>
                    {round}
                  </option>
                ))}
              </select>
            </FormField>
            <FormField label="FAAB budget" help="Annual free-agent acquisition budget for each franchise.">
              <input
                type="number"
                min="0"
                step="1"
                value={rules.faabBudget}
                onChange={(event) => {
                  const faabBudget = Number(event.target.value);
                  setRules({ ...rules, faabBudget, faabTrades: faabBudget > 0 && rules.faabTrades });
                }}
              />
            </FormField>
            {rules.faabBudget > 0 ? (
              <label className="checkbox-label">
                <input
                  type="checkbox"
                  checked={rules.faabTrades}
                  onChange={(event) => setRules({ ...rules, faabTrades: event.target.checked })}
                />
                <span>Allow FAAB to be traded</span>
              </label>
            ) : null}
          </>
        ) : null}
        {rules.draftType === "auction" ? <AuctionSettings rules={rules} setRules={setRules} /> : null}
      </div>
      <DraftOrderEditor rules={rules} setRules={setRules} disabled={disabled} />
    </fieldset>
  );
}

function DraftOrderEditor({ rules, setRules, disabled }: SettingsProps) {
  if (rules.draftType === "auction") return null;
  return (
    <details className="draft-order-editor">
      <summary>Customize the full draft order</summary>
      <p className="field-help">
        Franchise identity stays permanent when its draft position changes. Choosing a franchise swaps it with the
        franchise already in that position.
      </p>
      <div className="form-grid">
        {rules.draftOrder.map((teamNumber, index) => (
          <FormField key={index} label={`Pick ${index + 1} franchise`}>
            <select
              value={teamNumber}
              disabled={disabled}
              onChange={(event) => {
                const selectedTeam = Number(event.target.value);
                const updated = moveTeamToPosition(rules.draftOrder, selectedTeam, index + 1);
                setRules({
                  ...rules,
                  draftOrder: updated,
                  draftPosition: updated.indexOf(rules.userTeamNumber) + 1,
                });
              }}
            >
              {rules.teamNames.map((name, teamIndex) => (
                <option key={teamIndex + 1} value={teamIndex + 1}>
                  {name || `Team ${teamIndex + 1}`}
                  {teamIndex + 1 === rules.userTeamNumber ? " (you)" : ""}
                </option>
              ))}
            </select>
          </FormField>
        ))}
      </div>
    </details>
  );
}

export function TeamSettings({ rules, setRules, disabled }: SettingsProps) {
  const ownTeamIndex = rules.userTeamNumber - 1;
  const updateTeamName = (index: number, value: string) =>
    setRules((current) => ({
      ...current,
      teamNames: current.teamNames.map((teamName, teamIndex) => (teamIndex === index ? value : teamName)),
    }));

  return (
    <fieldset disabled={disabled}>
      <legend>Team settings</legend>
      <p className="field-help">
        Team names are optional. Leave any name blank and DraftMeld will use My Team or Team 1, Team 2, and so on. You
        can rename teams later without changing their picks.
      </p>
      <div className="form-grid team-name-grid">
        <FormField label="Your team name">
          <input
            maxLength={80}
            value={rules.teamNames[ownTeamIndex] ?? ""}
            placeholder="My Team"
            onChange={(event) => updateTeamName(ownTeamIndex, event.target.value)}
          />
        </FormField>
      </div>
      <details className="opponent-name-settings">
        <summary>Name the other franchises (optional)</summary>
        <p className="field-help">You can leave every opponent blank and identify teams by number during the draft.</p>
        <div className="form-grid team-name-grid">
          {rules.teamNames.map((name, index) =>
            index === ownTeamIndex ? null : (
              <FormField key={index} label={`Franchise ${index + 1}`}>
                <input
                  maxLength={80}
                  value={name}
                  placeholder={`Team ${index + 1}`}
                  onChange={(event) => updateTeamName(index, event.target.value)}
                />
              </FormField>
            ),
          )}
        </div>
      </details>
    </fieldset>
  );
}

function AuctionSettings({ rules, setRules }: { rules: LeagueRules; setRules: SetRules }) {
  const update = (values: Partial<LeagueRules>) => setRules((current) => ({ ...current, ...values }));
  return (
    <>
      <FormField label="Team auction budget">
        <input
          type="number"
          min="1"
          step="1"
          required
          value={rules.auctionBudget}
          onChange={(event) => update({ auctionBudget: Number(event.target.value) })}
        />
      </FormField>
      <FormField label="Minimum bid">
        <input
          type="number"
          min="1"
          max={rules.auctionBudget}
          step="1"
          required
          value={rules.auctionMinimumBid}
          onChange={(event) => update({ auctionMinimumBid: Number(event.target.value) })}
        />
      </FormField>
      <FormField label="My keeper spend" help="Dollars already committed by your team.">
        <input
          type="number"
          min="0"
          max={rules.auctionBudget}
          step="1"
          value={rules.myKeeperSpend}
          onChange={(event) => update({ myKeeperSpend: Number(event.target.value) })}
        />
      </FormField>
      <FormField label="League-wide keeper spend" help="Total dollars already committed to keepers across every team.">
        <input
          type="number"
          min={rules.myKeeperSpend}
          max={rules.auctionBudget * rules.teamCount}
          step="1"
          value={rules.keeperBudgetSpent}
          onChange={(event) => update({ keeperBudgetSpent: Number(event.target.value) })}
        />
      </FormField>
      <FormField label="Keeper value removed" help="DraftMeld baseline dollar value of players already kept.">
        <input
          type="number"
          min="0"
          step="1"
          value={rules.keeperValueRemoved}
          onChange={(event) => update({ keeperValueRemoved: Number(event.target.value) })}
        />
      </FormField>
      <label className="checkbox-label">
        <input
          type="checkbox"
          checked={rules.auctionBudgetTrades}
          onChange={(event) => update({ auctionBudgetTrades: event.target.checked })}
        />
        <span>Allow auction budget to be traded</span>
      </label>
    </>
  );
}
