import type { Dispatch, SetStateAction } from "react";
import type { LeagueRules } from "../shared/api/types";
import { FormField } from "../shared/ui/FormField";
import { LeagueRulesImport } from "./LeagueRulesImport";
import type { OnboardingImportReview } from "./onboarding";

interface StepProps {
  rules: LeagueRules;
  setRules: Dispatch<SetStateAction<LeagueRules>>;
}

export function LeagueBasicsStep({
  rules,
  setRules,
  importReview,
  setImportReview,
}: StepProps & {
  importReview?: OnboardingImportReview;
  setImportReview: Dispatch<SetStateAction<OnboardingImportReview | undefined>>;
}) {
  function updateTeamCount(teamCount: number) {
    const safeCount = Math.max(2, Math.min(32, teamCount));
    const userTeamNumber = Math.min(rules.userTeamNumber, safeCount);
    setRules({
      ...rules,
      teamCount: safeCount,
      userTeamNumber,
      draftPosition: Math.min(rules.draftPosition, safeCount),
      teamNames: Array.from({ length: safeCount }, (_, index) => rules.teamNames[index] ?? ""),
      draftOrder: Array.from({ length: safeCount }, (_, index) => index + 1),
    });
  }

  return (
    <section aria-labelledby="league-basics-title">
      <h2 id="league-basics-title" tabIndex={-1}>
        League basics
      </h2>
      <p>Start with the few details DraftMeld needs. Every setting can be changed later.</p>
      <LeagueRulesImport
        rules={rules}
        setRules={setRules}
        importReview={importReview}
        setImportReview={setImportReview}
      />
      <div className="form-grid">
        <FormField label="League name">
          <input required value={rules.name} onChange={(event) => setRules({ ...rules, name: event.target.value })} />
        </FormField>
        <FormField label="Number of teams">
          <input
            required
            type="number"
            min="2"
            max="32"
            value={rules.teamCount}
            onChange={(event) => updateTeamCount(Number(event.target.value))}
          />
        </FormField>
        <FormField label="Your draft position" help="You do not need to know the other team names.">
          <select
            value={rules.draftPosition}
            onChange={(event) =>
              setRules({
                ...rules,
                draftPosition: Number(event.target.value),
                userTeamNumber: Number(event.target.value),
              })
            }
          >
            {Array.from({ length: rules.teamCount }, (_, index) => (
              <option key={index + 1} value={index + 1}>
                Pick {index + 1}
              </option>
            ))}
          </select>
        </FormField>
        <FormField label="League format">
          <select
            value={rules.leagueFormat}
            onChange={(event) => {
              const leagueFormat = event.target.value as LeagueRules["leagueFormat"];
              setRules({ ...rules, leagueFormat, futurePickSeasons: leagueFormat === "dynasty" ? 3 : 0 });
            }}
          >
            <option value="redraft">Redraft</option>
            <option value="dynasty">Dynasty</option>
          </select>
        </FormField>
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
      </div>
    </section>
  );
}

export function OnboardingReviewStep({ rules }: { rules: LeagueRules }) {
  const activeScoringRules = Object.values(rules.scoringRules).filter((value) => value !== 0).length;
  const startingSlots = rules.rosterSlots
    .filter((slot) => slot.isStarting)
    .map((slot) => `${slot.count} ${slot.name}`)
    .join(", ");
  const reserveSlots = rules.rosterSlots
    .filter((slot) => !slot.isStarting)
    .reduce((total, slot) => total + slot.count, 0);
  return (
    <section aria-labelledby="review-step-title">
      <h2 id="review-step-title" tabIndex={-1}>
        Review and create
      </h2>
      <p>Nothing is created until you choose Create league.</p>
      <dl className="setup-review">
        <div>
          <dt>League</dt>
          <dd>{rules.name}</dd>
        </div>
        <div>
          <dt>Teams</dt>
          <dd>{rules.teamCount}</dd>
        </div>
        <div>
          <dt>Draft</dt>
          <dd>
            {rules.draftType} from pick {rules.draftPosition}
          </dd>
        </div>
        <div>
          <dt>League format</dt>
          <dd>{rules.leagueFormat}</dd>
        </div>
        <div>
          <dt>Starting roster</dt>
          <dd>{startingSlots}</dd>
        </div>
        <div>
          <dt>Reserve slots</dt>
          <dd>{reserveSlots}</dd>
        </div>
        {rules.draftType === "auction" ? (
          <div>
            <dt>Auction budget</dt>
            <dd>
              ${rules.auctionBudget} with a ${rules.auctionMinimumBid} minimum bid
            </dd>
          </div>
        ) : null}
        {rules.leagueFormat === "dynasty" ? (
          <div>
            <dt>Dynasty assets</dt>
            <dd>
              {rules.rookieDraftRounds} rookie rounds and {rules.futurePickSeasons} future pick seasons
            </dd>
          </div>
        ) : null}
        <div>
          <dt>Reception scoring</dt>
          <dd>{rules.scoringRules.reception ?? 0} points per reception</dd>
        </div>
        <div>
          <dt>Active scoring values</dt>
          <dd>{activeScoringRules}</dd>
        </div>
      </dl>
      <p className="field-help">After creation, DraftMeld will guide you to ranking sources before the draft board.</p>
    </section>
  );
}
