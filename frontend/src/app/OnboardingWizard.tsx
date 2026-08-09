import { useEffect, useRef, useState } from "react";
import type { League, LeagueRules } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { Panel } from "../shared/ui/Panel";
import { StatusMessage } from "../shared/ui/StatusMessage";
import { LeagueForm } from "./LeagueForm";
import {
  completeOnboarding,
  loadOnboardingDraft,
  newOnboardingDraft,
  saveOnboardingDraft,
  type OnboardingImportReview,
} from "./onboarding";
import { OnboardingScoringStep } from "./OnboardingScoringStep";
import { LeagueBasicsStep, OnboardingReviewStep } from "./OnboardingSteps";

const steps = ["League basics", "Scoring", "Review"] as const;
const stepHeadingIds = ["league-basics-title", "scoring-step-title", "review-step-title"] as const;

interface OnboardingWizardProps {
  onClose: (message?: string) => void;
  onCreate: (rules: LeagueRules) => Promise<League>;
  onOpenDraft: (leagueId: string) => void;
  onOpenRankings: (leagueId: string) => void;
}

export function OnboardingWizard({ onClose, onCreate, onOpenDraft, onOpenRankings }: OnboardingWizardProps) {
  const [initial] = useState(() => loadOnboardingDraft() ?? newOnboardingDraft());
  const [step, setStep] = useState(initial.step);
  const [rules, setRules] = useState(initial.rules);
  const [importReview, setImportReview] = useState<OnboardingImportReview | undefined>(initial.importReview);
  const [manual, setManual] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [created, setCreated] = useState<League | null>(null);
  const heading = useRef<HTMLHeadingElement>(null);
  const initialFocusSet = useRef(false);

  useEffect(() => {
    if (!created) saveOnboardingDraft(step, rules, importReview);
  }, [created, importReview, rules, step]);

  useEffect(() => {
    if (manual || created || (!initialFocusSet.current && step === 0)) {
      heading.current?.focus();
    } else {
      document.getElementById(stepHeadingIds[step])?.focus();
    }
    initialFocusSet.current = true;
  }, [created, manual, step]);

  async function create(rulesToSave: LeagueRules) {
    setBusy(true);
    setError("");
    try {
      const league = await onCreate(rulesToSave);
      completeOnboarding();
      setCreated(league);
      setManual(false);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "The league could not be created.");
    } finally {
      setBusy(false);
    }
  }

  function skipGuide() {
    completeOnboarding();
    onClose();
  }

  function saveAndClose() {
    onClose("Setup progress saved. Resume the setup guide whenever you are ready.");
  }

  if (created) {
    return (
      <main className="onboarding-layout">
        <Panel variant="form" className="onboarding-panel" aria-labelledby="onboarding-complete-title">
          <p className="eyebrow">Setup complete</p>
          <h1 id="onboarding-complete-title" ref={heading} tabIndex={-1}>
            {created.name} is ready
          </h1>
          <p>
            Your league is saved. Ranking sources are the recommended next step; projections remain optional for a
            simpler setup.
          </p>
          <div className="onboarding-next-actions">
            <Button variant="primary" onClick={() => onOpenRankings(created.id)}>
              Configure ranking sources
            </Button>
            <Button onClick={() => onOpenDraft(created.id)}>Open draft board</Button>
            <Button onClick={() => onClose()}>Finish for now</Button>
          </div>
        </Panel>
      </main>
    );
  }

  if (manual) {
    return (
      <main className="onboarding-layout">
        <Panel variant="form" className="onboarding-panel" aria-labelledby="manual-setup-title">
          <div className="onboarding-utility-actions">
            <Button onClick={saveAndClose}>Save and finish later</Button>
            <Button variant="dangerText" onClick={skipGuide}>
              Skip setup guide
            </Button>
          </div>
          <h1 id="manual-setup-title" ref={heading} tabIndex={-1}>
            Configure every league setting
          </h1>
          <LeagueForm
            initialRules={rules}
            busy={busy}
            onCancel={() => setManual(false)}
            onChange={setRules}
            onSave={create}
          />
          {error ? <StatusMessage tone="error">{error}</StatusMessage> : null}
        </Panel>
      </main>
    );
  }

  return (
    <main className="onboarding-layout">
      <Panel variant="form" className="onboarding-panel" aria-labelledby="onboarding-title">
        <div className="form-heading">
          <div>
            <p className="eyebrow">First-run guide</p>
            <h1 id="onboarding-title" ref={heading} tabIndex={-1}>
              Set up your draft workspace
            </h1>
            <p>Your progress is saved on this device after every step.</p>
          </div>
          <div className="onboarding-utility-actions">
            <Button onClick={saveAndClose}>Save and finish later</Button>
            <Button onClick={() => setManual(true)}>Configure manually</Button>
            <Button variant="dangerText" onClick={skipGuide}>
              Skip setup guide
            </Button>
          </div>
        </div>

        <ol className="wizard-progress" aria-label="Setup progress">
          {steps.map((label, index) => (
            <li key={label} aria-current={step === index ? "step" : undefined}>
              <span>{index + 1}</span> {label}
            </li>
          ))}
        </ol>

        <div hidden={step !== 0}>
          <LeagueBasicsStep
            rules={rules}
            setRules={setRules}
            importReview={importReview}
            setImportReview={setImportReview}
          />
        </div>
        <div hidden={step !== 1}>
          <OnboardingScoringStep rules={rules} setRules={setRules} />
        </div>
        <div hidden={step !== 2}>
          <OnboardingReviewStep rules={rules} />
        </div>
        {error ? <StatusMessage tone="error">{error}</StatusMessage> : null}

        <div className="form-actions wizard-actions">
          {step > 0 ? <Button onClick={() => setStep((current) => current - 1)}>Back</Button> : null}
          {step < steps.length - 1 ? (
            <Button variant="primary" onClick={() => setStep((current) => current + 1)}>
              Continue
            </Button>
          ) : (
            <Button variant="primary" disabled={busy} onClick={() => create(rules)}>
              {busy ? "Creating league..." : "Create league"}
            </Button>
          )}
        </div>
      </Panel>
    </main>
  );
}
