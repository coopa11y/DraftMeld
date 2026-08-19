import { useEffect, useRef, useState } from "react";
import { leagueToRules } from "../shared/api/leagues";
import type { League, LeagueRules } from "../shared/api/types";
import { useViewHeadingFocus } from "../shared/hooks/useViewHeadingFocus";
import { Button } from "../shared/ui/Button";
import { Dialog } from "../shared/ui/Dialog";
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
  onSave: (rules: LeagueRules) => Promise<League | void>;
  onSaveAndExit?: (rules: LeagueRules) => Promise<League | void>;
}

const sections = ["League", "Draft", "Teams", "Roster", "Scoring"] as const;
type Section = (typeof sections)[number];

export function LeagueForm({ league, initialRules, busy, onCancel, onChange, onSave, onSaveAndExit }: LeagueFormProps) {
  const [rules, setRules] = useState<LeagueRules>(() =>
    league ? leagueToRules(league) : initialRules ? cloneLeagueRules(initialRules) : defaultLeagueRules(),
  );
  const [savedRules, setSavedRules] = useState<LeagueRules>(() => cloneLeagueRules(rules));
  const [section, setSection] = useState<Section>("League");
  const [discardOpen, setDiscardOpen] = useState(false);
  const heading = useViewHeadingFocus<HTMLHeadingElement>();
  const discardHeading = useRef<HTMLHeadingElement>(null);
  const hasUnsavedChanges = JSON.stringify(rules) !== JSON.stringify(savedRules);

  useEffect(() => {
    onChange?.(rules);
  }, [onChange, rules]);

  async function handleSave(action: (rules: LeagueRules) => Promise<League | void>, exit: boolean) {
    try {
      const saved = await action(rules);
      if (!exit) {
        const normalized = saved ? leagueToRules(saved) : cloneLeagueRules(rules);
        setRules(normalized);
        setSavedRules(cloneLeagueRules(normalized));
      }
    } catch {
      // The parent presents the API error while this form preserves the user's changes.
    }
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

      <p className="field-help">{hasUnsavedChanges ? "Unsaved changes" : "All changes saved"}</p>

      <div className="form-actions">
        <Button
          variant="primary"
          disabled={busy}
          onClick={(event) => {
            if (event.currentTarget.form?.reportValidity() === false) return;
            void handleSave(onSave, false);
          }}
        >
          {busy ? "Saving…" : onSaveAndExit ? "Save changes" : "Save league"}
        </Button>
        {onSaveAndExit ? (
          <Button
            disabled={busy}
            onClick={(event) => {
              if (event.currentTarget.form?.reportValidity() === false) return;
              void handleSave(onSaveAndExit, true);
            }}
          >
            Save and exit
          </Button>
        ) : null}
        <Button onClick={() => (hasUnsavedChanges ? setDiscardOpen(true) : onCancel())} disabled={busy}>
          Cancel
        </Button>
      </div>
      <Dialog
        open={discardOpen}
        labelledBy="discard-league-changes-title"
        initialFocusRef={discardHeading}
        onClose={() => setDiscardOpen(false)}
      >
        <h2 id="discard-league-changes-title" ref={discardHeading} tabIndex={-1}>
          Discard unsaved changes?
        </h2>
        <p>Your changes across every league configuration section will be lost.</p>
        <div className="dialog-actions">
          <Button variant="danger" onClick={onCancel}>
            Discard changes
          </Button>
          <Button onClick={() => setDiscardOpen(false)}>Keep editing</Button>
        </div>
      </Dialog>
    </form>
  );
}
