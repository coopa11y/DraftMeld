import { useEffect, useRef, useState } from "react";
import { createLeague, listLeagues } from "../shared/api/leagues";
import type { League, LeagueRules } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { StatusMessage } from "../shared/ui/StatusMessage";
import { DraftWorkspace } from "./DraftWorkspace";
import { LeagueManager } from "./LeagueManager";
import { OnboardingWizard } from "./OnboardingWizard";
import { RankingSources } from "./RankingSources";
import { completeOnboarding, loadOnboardingDraft, onboardingComplete } from "./onboarding";

const ACTIVE_LEAGUE_KEY = "draftmeld.active-league.v1";

export function App() {
  const [leagues, setLeagues] = useState<League[]>([]);
  const [activeLeagueId, setActiveLeagueId] = useState(() => localStorage.getItem(ACTIVE_LEAGUE_KEY) ?? "demo");
  const initialActiveLeagueId = useRef(activeLeagueId);
  const [view, setView] = useState<"draft" | "leagues" | "rankings">("draft");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [onboardingOpen, setOnboardingOpen] = useState(false);
  const [onboardingDone, setOnboardingDone] = useState(onboardingComplete);
  const [hasOnboardingDraft, setHasOnboardingDraft] = useState(() => loadOnboardingDraft() !== null);

  useEffect(() => {
    let active = true;
    listLeagues()
      .then((loaded) => {
        if (!active) return;
        setLeagues(loaded);
        const selected = loaded.some((league) => league.id === initialActiveLeagueId.current)
          ? initialActiveLeagueId.current
          : loaded[0]?.id;
        if (selected) selectLeague(selected);
        else setView("leagues");
      })
      .catch((reason: Error) => {
        if (active) setError(reason.message);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  function selectLeague(id: string) {
    setActiveLeagueId(id);
    localStorage.setItem(ACTIVE_LEAGUE_KEY, id);
  }

  function handleLeaguesChange(updated: League[], preferredId?: string) {
    setLeagues(updated);
    const nextId =
      preferredId ?? (updated.some((league) => league.id === activeLeagueId) ? activeLeagueId : updated[0]?.id);
    if (nextId) selectLeague(nextId);
    else setView("leagues");
  }

  function handleLeagueUpdated(updated: League) {
    setLeagues((current) => current.map((league) => (league.id === updated.id ? updated : league)));
  }

  async function handleOnboardingCreate(rules: LeagueRules): Promise<League> {
    const saved = await createLeague(rules);
    setLeagues((current) => [...current.filter((league) => league.id !== saved.id), saved]);
    selectLeague(saved.id);
    setOnboardingDone(true);
    setHasOnboardingDraft(false);
    return saved;
  }

  function closeOnboarding() {
    setOnboardingOpen(false);
    setOnboardingDone(onboardingComplete());
    setHasOnboardingDraft(loadOnboardingDraft() !== null);
  }

  const activeLeague = leagues.find((league) => league.id === activeLeagueId);

  if (loading) {
    return (
      <main className="centered-status" aria-busy="true">
        <StatusMessage>Loading leagues...</StatusMessage>
      </main>
    );
  }

  return (
    <>
      <header className="app-header">
        <div>
          <a className="brand" href="/" aria-label="DraftMeld home">
            DraftMeld
          </a>
          <span className="version">v{__APP_VERSION__}</span>
        </div>
        <nav className="league-navigation" aria-label="DraftMeld navigation">
          {leagues.length > 0 ? (
            <label>
              <span>Active league</span>
              <select
                value={activeLeagueId}
                onChange={(event) => {
                  selectLeague(event.target.value);
                  setView("draft");
                }}
              >
                {leagues.map((league) => (
                  <option key={league.id} value={league.id}>
                    {league.name}
                  </option>
                ))}
              </select>
            </label>
          ) : null}
          {view !== "leagues" ? <Button onClick={() => setView("leagues")}>Manage leagues</Button> : null}
          {view !== "rankings" ? (
            <Button onClick={() => setView("rankings")} disabled={!activeLeague}>
              Ranking sources
            </Button>
          ) : null}
          {view !== "draft" ? (
            <Button onClick={() => setView("draft")} disabled={leagues.length === 0}>
              Return to draft
            </Button>
          ) : null}
          {!onboardingDone && !onboardingOpen ? (
            <Button onClick={() => setOnboardingOpen(true)}>
              {hasOnboardingDraft ? "Resume setup" : "Setup guide"}
            </Button>
          ) : null}
        </nav>
      </header>

      {error ? <StatusMessage tone="error">Unable to load leagues. {error}</StatusMessage> : null}
      {!onboardingDone && !onboardingOpen ? (
        <section className="onboarding-prompt" aria-labelledby="onboarding-prompt-title">
          <div>
            <p className="eyebrow">Getting started</p>
            <h2 id="onboarding-prompt-title">
              {hasOnboardingDraft ? "Your saved setup is ready" : "Set up your league"}
            </h2>
            <p>
              {hasOnboardingDraft
                ? "Resume where you stopped. Your saved league settings remain on this device."
                : "Use the guided setup, import league scoring rules, or configure everything manually."}
            </p>
          </div>
          <div className="onboarding-prompt-actions">
            <Button variant="primary" onClick={() => setOnboardingOpen(true)}>
              {hasOnboardingDraft ? "Resume setup" : "Start setup"}
            </Button>
            <Button
              onClick={() => {
                completeOnboarding();
                setOnboardingDone(true);
                setHasOnboardingDraft(false);
              }}
            >
              Dismiss guide
            </Button>
          </div>
        </section>
      ) : null}
      {onboardingOpen ? (
        <OnboardingWizard
          onClose={closeOnboarding}
          onCreate={handleOnboardingCreate}
          onOpenDraft={(id) => {
            selectLeague(id);
            setView("draft");
            setOnboardingOpen(false);
          }}
          onOpenRankings={(id) => {
            selectLeague(id);
            setView("rankings");
            setOnboardingOpen(false);
          }}
        />
      ) : view === "rankings" && activeLeague ? (
        <RankingSources key={activeLeague.id} league={activeLeague} onLeagueUpdated={handleLeagueUpdated} />
      ) : view === "leagues" ? (
        <LeagueManager
          leagues={leagues}
          activeLeagueId={activeLeagueId}
          onLeaguesChange={handleLeaguesChange}
          onOpenDraft={(id) => {
            selectLeague(id);
            setView("draft");
          }}
        />
      ) : (
        <DraftWorkspace key={activeLeagueId} leagueId={activeLeagueId} />
      )}
    </>
  );
}
