import type { Dispatch, SetStateAction } from "react";
import type { LeagueRules } from "../shared/api/types";
import { FormField } from "../shared/ui/FormField";

type SetRules = Dispatch<SetStateAction<LeagueRules>>;

function resizeTeamNames(names: string[], teamCount: number): string[] {
  return Array.from({ length: teamCount }, (_, index) => names[index] ?? "");
}

function moveUserTeam(names: string[], previousPosition: number, nextPosition: number): string[] {
  const updated = [...names];
  [updated[previousPosition - 1], updated[nextPosition - 1]] = [
    updated[nextPosition - 1] ?? "",
    updated[previousPosition - 1] ?? "",
  ];
  return updated;
}

interface SettingsProps {
  rules: LeagueRules;
  setRules: SetRules;
  disabled: boolean;
}

export function LeagueSettings({ rules, setRules, disabled }: SettingsProps) {
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
              const draftPosition = Math.min(rules.draftPosition, teamCount);
              setRules({
                ...rules,
                teamCount,
                draftPosition,
                teamNames: resizeTeamNames(
                  moveUserTeam(rules.teamNames, rules.draftPosition, draftPosition),
                  teamCount,
                ),
              });
            }}
          />
        </FormField>
        <FormField label="Scoring preset">
          <select
            value={scoringPreset(rules.scoringRules.reception)}
            onChange={(event) => {
              if (event.target.value === "custom") return;
              setRules({ ...rules, scoringRules: { ...rules.scoringRules, reception: Number(event.target.value) } });
            }}
          >
            <option value="0">Standard</option>
            <option value="0.5">Half PPR</option>
            <option value="1">PPR</option>
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
              setRules({
                ...rules,
                draftPosition,
                teamNames: moveUserTeam(rules.teamNames, rules.draftPosition, draftPosition),
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
        {rules.draftType === "auction" ? <AuctionSettings rules={rules} setRules={setRules} /> : null}
      </div>
    </fieldset>
  );
}

export function TeamSettings({ rules, setRules, disabled }: SettingsProps) {
  return (
    <fieldset disabled={disabled}>
      <legend>Team settings</legend>
      <p className="field-help">
        Team names are optional. Leave any name blank and DraftMeld will use My Team or Team 1, Team 2, and so on. You
        can rename teams later without changing their picks.
      </p>
      <div className="form-grid team-name-grid">
        {rules.teamNames.map((name, index) => (
          <FormField key={index} label={`Team ${index + 1}${index + 1 === rules.draftPosition ? " (your team)" : ""}`}>
            <input
              maxLength={80}
              value={name}
              placeholder={index + 1 === rules.draftPosition ? "My Team" : `Team ${index + 1}`}
              onChange={(event) =>
                setRules((current) => ({
                  ...current,
                  teamNames: current.teamNames.map((teamName, teamIndex) =>
                    teamIndex === index ? event.target.value : teamName,
                  ),
                }))
              }
            />
          </FormField>
        ))}
      </div>
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
    </>
  );
}

function scoringPreset(receptions: number): string {
  return receptions === 0 || receptions === 0.5 || receptions === 1 ? String(receptions) : "custom";
}
