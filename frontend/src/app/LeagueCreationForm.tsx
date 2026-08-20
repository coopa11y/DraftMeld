import { useState, type FormEvent } from "react";
import type { LeagueRules } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { FormField } from "../shared/ui/FormField";
import {
  DraftSettings,
  LeagueSettings,
  moveTeamToPosition,
  resizeDraftOrder,
  resizeTeamNames,
  TeamSettings,
} from "./LeagueSetupSections";
import { LeagueRulesImport } from "./LeagueRulesImport";
import { RosterSettings } from "./RosterSettings";
import { ScoringSettings } from "./ScoringSettings";
import { defaultLeagueRules } from "./leagueDefaults";
import { enabledRosterPositions } from "./rosterConfiguration";
import { applyReceptionPreset, receptionPreset } from "./scoring";
import type { OnboardingImportReview } from "./onboarding";

interface LeagueCreationFormProps {
  busy: boolean;
  onCancel: () => void;
  onSave: (rules: LeagueRules) => Promise<void>;
}

const sections = ["League", "Draft", "Teams", "Roster", "Scoring"] as const;
type Section = (typeof sections)[number];

export function LeagueCreationForm({ busy, onCancel, onSave }: LeagueCreationFormProps) {
  const [rules, setRules] = useState(defaultLeagueRules);
  const [advancedOpen, setAdvancedOpen] = useState(false);
  const [section, setSection] = useState<Section>("League");
  const [importReview, setImportReview] = useState<OnboardingImportReview>();

  async function submit(event: FormEvent) {
    event.preventDefault();
    await onSave(rules);
  }

  function updateTeamCount(teamCount: number) {
    const userTeamNumber = Math.min(rules.userTeamNumber, teamCount);
    const draftOrder = resizeDraftOrder(rules.draftOrder, teamCount);
    setRules({
      ...rules,
      teamCount,
      userTeamNumber,
      draftOrder,
      draftPosition: rules.draftPosition === 0 ? 0 : draftOrder.indexOf(userTeamNumber) + 1,
      teamNames: resizeTeamNames(rules.teamNames, teamCount),
    });
  }

  return (
    <div className="league-form league-creation-form">
      <form onSubmit={submit}>
        <div className="form-heading">
          <div>
            <p className="eyebrow">League setup</p>
            <h1>Create a league</h1>
            <p>Start with the essentials. You can open advanced settings now or change any setting later.</p>
          </div>
        </div>

        <fieldset disabled={busy} className="basic-league-setup">
          <legend>Basic setup</legend>
          <LeagueRulesImport
            headingLevel={2}
            rules={rules}
            setRules={setRules}
            importReview={importReview}
            setImportReview={setImportReview}
          />
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
              <select value={rules.teamCount} onChange={(event) => updateTeamCount(Number(event.target.value))}>
                {[8, 10, 12, 14, 16].map((count) => (
                  <option key={count} value={count}>
                    {count}
                  </option>
                ))}
              </select>
            </FormField>
            <FormField label="Scoring">
              <select
                value={receptionPreset(rules.scoringRules)}
                onChange={(event) => {
                  if (event.target.value !== "custom")
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
            <FormField label="Draft type">
              <select
                value={rules.draftType}
                onChange={(event) => setRules({ ...rules, draftType: event.target.value as LeagueRules["draftType"] })}
              >
                <option value="snake">Snake</option>
                <option value="linear">Linear</option>
                <option value="auction">Auction</option>
              </select>
            </FormField>
            {rules.draftType !== "auction" ? (
              <FormField
                label="Your draft position"
                help="Leave this not set if your league has not assigned your pick yet."
              >
                <select
                  value={rules.draftPosition}
                  onChange={(event) => {
                    const draftPosition = Number(event.target.value);
                    if (draftPosition === 0) {
                      setRules({ ...rules, draftPosition: 0 });
                      return;
                    }
                    setRules({
                      ...rules,
                      draftPosition,
                      draftOrder: moveTeamToPosition(rules.draftOrder, rules.userTeamNumber, draftPosition),
                    });
                  }}
                >
                  <option value="0">Not set yet</option>
                  {Array.from({ length: rules.teamCount }, (_, index) => (
                    <option key={index + 1} value={index + 1}>
                      Pick {index + 1}
                    </option>
                  ))}
                </select>
              </FormField>
            ) : null}
          </div>
        </fieldset>

        <Button
          className="advanced-settings-toggle"
          aria-expanded={advancedOpen}
          aria-controls="advanced-league-settings"
          onClick={() => setAdvancedOpen((open) => !open)}
        >
          {advancedOpen ? "Hide advanced settings" : "Advanced settings"}
        </Button>

        {advancedOpen ? (
          <section
            id="advanced-league-settings"
            className="advanced-league-settings"
            aria-labelledby="advanced-settings-title"
          >
            <h2 id="advanced-settings-title">Advanced settings</h2>
            <p>Configure league format, teams, roster positions, detailed scoring, auction rules, and draft order.</p>
            <nav className="league-form-nav" aria-label="Advanced league configuration sections">
              {sections.map((item) => (
                <Button
                  key={item}
                  className="league-form-nav-button"
                  aria-pressed={section === item}
                  onClick={() => setSection(item)}
                  disabled={busy}
                >
                  {item}
                </Button>
              ))}
            </nav>
            <div className="league-form-section" aria-label={`${section} advanced configuration`}>
              {section === "League" ? (
                <LeagueSettings rules={rules} setRules={setRules} disabled={busy} hideBasic />
              ) : null}
              {section === "Draft" ? (
                <DraftSettings rules={rules} setRules={setRules} disabled={busy} hideBasic />
              ) : null}
              {section === "Teams" ? <TeamSettings rules={rules} setRules={setRules} disabled={busy} /> : null}
              {section === "Roster" ? <RosterSettings rules={rules} setRules={setRules} disabled={busy} /> : null}
              {section === "Scoring" ? (
                <ScoringSettings
                  rules={rules}
                  setRules={setRules}
                  disabled={busy}
                  enabledPositions={enabledRosterPositions(rules.rosterSlots)}
                />
              ) : null}
            </div>
          </section>
        ) : null}

        <div className="form-actions">
          <Button type="submit" variant="primary" disabled={busy || !rules.name.trim()}>
            {busy ? "Saving…" : "Create league"}
          </Button>
          <Button disabled={busy} onClick={onCancel}>
            Cancel
          </Button>
        </div>
      </form>
    </div>
  );
}
