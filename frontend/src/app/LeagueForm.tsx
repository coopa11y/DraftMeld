import { useEffect, useState } from "react";
import { leagueToRules } from "../shared/api/leagues";
import type { League, LeagueRules } from "../shared/api/types";
import { useViewHeadingFocus } from "../shared/hooks/useViewHeadingFocus";
import { Button } from "../shared/ui/Button";
import { DraftSettings, LeagueSettings, TeamSettings } from "./LeagueSetupSections";
import { RosterSettings } from "./RosterSettings";
import { ScoringSettings } from "./ScoringSettings";
import { cloneLeagueRules, defaultLeagueRules } from "./leagueDefaults";
import { enabledRosterPositions } from "./rosterConfiguration";

interface LeagueFormProps {
  league?: League;
  initialRules?: LeagueRules;
  busy: boolean;
  onCancel: () => void;
  onChange?: (rules: LeagueRules) => void;
  onSave: (rules: LeagueRules) => Promise<void>;
}

const sections = ["League", "Draft", "Teams", "Roster", "Scoring"] as const;
type Section = (typeof sections)[number];

export function LeagueForm({ league, initialRules, busy, onCancel, onChange, onSave }: LeagueFormProps) {
  const [rules, setRules] = useState<LeagueRules>(() =>
    league ? leagueToRules(league) : initialRules ? cloneLeagueRules(initialRules) : defaultLeagueRules(),
  );
  const [section, setSection] = useState<Section>("League");
  const heading = useViewHeadingFocus<HTMLHeadingElement>();

  useEffect(() => {
    onChange?.(rules);
  }, [onChange, rules]);

  async function handleSave() {
    await onSave(rules);
  }

  return (
    <form className="league-form" onSubmit={(event) => event.preventDefault()}>
      <div className="form-heading">
        <div>
          <p className="eyebrow">League setup</p>
          <h1 ref={heading} tabIndex={-1}>
            {league ? `Edit ${league.name}` : "Create a league"}
          </h1>
          <p>Configure one section at a time. Your changes stay in place as you move between sections.</p>
        </div>
      </div>

      <nav className="league-form-nav" aria-label="League configuration sections">
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

      <section className="league-form-section" aria-label={`${section} configuration`}>
        {section === "League" ? (
          <LeagueSettings
            rules={rules}
            setRules={setRules}
            disabled={busy}
            lockSeason={Boolean(league && rules.leagueFormat === "dynasty")}
          />
        ) : null}
        {section === "Draft" ? <DraftSettings rules={rules} setRules={setRules} disabled={busy} /> : null}
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
      </section>

      <div className="form-actions">
        <Button
          variant="primary"
          disabled={busy}
          onClick={(event) => {
            if (event.currentTarget.form?.reportValidity() === false) return;
            void handleSave();
          }}
        >
          {busy ? "Saving…" : "Save league"}
        </Button>
        <Button onClick={onCancel} disabled={busy}>
          Cancel
        </Button>
      </div>
    </form>
  );
}
