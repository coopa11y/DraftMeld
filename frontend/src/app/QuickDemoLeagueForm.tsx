import { useState, type FormEvent } from "react";
import type { LeagueRules } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { FormField } from "../shared/ui/FormField";
import { buildQuickDemoRules, defaultQuickDemoSettings, type QuickDemoSettings } from "./quickDemoRules";

export type QuickDemoDestination = "league" | "mock";

interface QuickDemoLeagueFormProps {
  busy: boolean;
  onCancel: () => void;
  onCreate: (rules: LeagueRules, destination: QuickDemoDestination) => Promise<void>;
}

export function QuickDemoLeagueForm({ busy, onCancel, onCreate }: QuickDemoLeagueFormProps) {
  const [settings, setSettings] = useState<QuickDemoSettings>(defaultQuickDemoSettings);
  const update = (values: Partial<QuickDemoSettings>) => setSettings((current) => ({ ...current, ...values }));

  function create(destination: QuickDemoDestination) {
    return onCreate(buildQuickDemoRules(settings), destination);
  }

  function submit(event: FormEvent) {
    event.preventDefault();
    void create("league");
  }

  return (
    <form className="quick-demo-form" aria-labelledby="quick-demo-title" onSubmit={submit}>
      <div>
        <h2 id="quick-demo-title">Quick demo league</h2>
        <p>Create a separate practice league with common defaults. Every setting can be changed later.</p>
      </div>
      <div className="form-grid quick-demo-grid">
        <FormField label="League name">
          <input
            required
            maxLength={80}
            value={settings.name}
            onChange={(event) => update({ name: event.target.value })}
          />
        </FormField>
        <FormField label="Teams">
          <select
            value={settings.teamCount}
            onChange={(event) => {
              const teamCount = Number(event.target.value);
              update({ teamCount, draftPosition: Math.min(settings.draftPosition, teamCount) });
            }}
          >
            {[8, 10, 12, 14, 16].map((count) => (
              <option key={count} value={count}>
                {count}
              </option>
            ))}
          </select>
        </FormField>
        {settings.draftType !== "auction" ? (
          <FormField label="Draft position">
            <select
              value={settings.draftPosition}
              onChange={(event) => update({ draftPosition: Number(event.target.value) })}
            >
              {Array.from({ length: settings.teamCount }, (_, index) => index + 1).map((position) => (
                <option key={position} value={position}>
                  Pick {position}
                </option>
              ))}
            </select>
          </FormField>
        ) : null}
        <FormField label="Scoring">
          <select
            value={settings.receptionPreset}
            onChange={(event) =>
              update({ receptionPreset: event.target.value as QuickDemoSettings["receptionPreset"] })
            }
          >
            <option value="0">Standard</option>
            <option value="0.5">Half PPR</option>
            <option value="1">PPR</option>
          </select>
        </FormField>
        <FormField label="Draft type">
          <select
            value={settings.draftType}
            onChange={(event) => update({ draftType: event.target.value as LeagueRules["draftType"] })}
          >
            <option value="snake">Snake</option>
            <option value="linear">Linear</option>
            <option value="auction">Auction</option>
          </select>
        </FormField>
      </div>
      <div className="form-actions">
        <Button type="submit" variant="primary" disabled={busy || !settings.name.trim()}>
          Create demo league
        </Button>
        <Button disabled={busy || !settings.name.trim()} onClick={() => void create("mock")}>
          Create and start mock draft
        </Button>
        <Button disabled={busy} onClick={onCancel}>
          Cancel
        </Button>
      </div>
    </form>
  );
}
