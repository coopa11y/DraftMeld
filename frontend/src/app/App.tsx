import { useEffect, useRef, useState } from "react";
import { createLeague, listLeagues } from "../shared/api/leagues";
import type { League, LeagueRules } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { StatusMessage } from "../shared/ui/StatusMessage";
import { DraftWorkspace } from "./DraftWorkspace";
import { LeagueManager } from "./LeagueManager";
import { OnboardingWizard } from "./OnboardingWizard";
import { RankingSources } from "./RankingSources";
import { loadOnboardingDraft, onboardingComplete } from "./onboarding";

const ACTIVE_LEAGUE_KEY = "draftmeld.active-league.v1";
const CREATE_LEAGUE_OPTION = "__create_league__";
const MANAGE_LEAGUES_OPTION = "__manage_leagues__";

export function App() {
  const [leagues, setLeagues] = useState<League[]>([]);
  const [activeLeagueId, setActiveLeagueId] = useState(() => localStorage.getItem(ACTIVE_LEAGUE_KEY) ?? "demo");
  const initialActiveLeagueId = useRef(activeLeagueId);
  const [view, setView] = useState<"draft" | "tools" | "leagues" | "rankings">("draft");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [onboardingOpen, setOnboardingOpen] = useState(false);
  const [onboardingDone, setOnboardingDone] = useState(onboardingComplete);
  const [hasOnboardingDraft, setHasOnboardingDraft] = useState(() => loadOnboardingDraft() !== null);
  const [notice, setNotice] = useState("");
  const noticeRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (notice) noticeRef.current?.focus();
  }, [notice]);

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

  function closeOnboarding(message?: string) {
    setOnboardingOpen(false);
    setOnboardingDone(onboardingComplete());
    setHasOnboardingDraft(loadOnboardingDraft() !== null);
    setNotice(message ?? "");
  }

  function openOnboarding() {
    setNotice("");
    setOnboardingOpen(true);
  }

  function handleLeagueSelection(value: string) {
    if (value === CREATE_LEAGUE_OPTION) {
      openOnboarding();
      return;
    }
    if (value === MANAGE_LEAGUES_OPTION) {
      setView("leagues");
      return;
    }
    selectLeague(value);
    setView("draft");
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
          <label>
            <span>Active league</span>
            <select value={activeLeague?.id ?? ""} onChange={(event) => handleLeagueSelection(event.target.value)}>
              {!activeLeague ? (
                <option value="" disabled>
                  No active league
                </option>
              ) : null}
              {leagues.map((league) => (
                <option key={league.id} value={league.id}>
                  {league.name}
                </option>
              ))}
              <optgroup label="League actions">
                <option value={CREATE_LEAGUE_OPTION}>Create a league…</option>
                <option value={MANAGE_LEAGUES_OPTION}>Manage leagues…</option>
              </optgroup>
            </select>
          </label>
          <Button
            aria-current={view === "draft" ? "page" : undefined}
            onClick={() => setView("draft")}
            disabled={leagues.length === 0}
          >
            Draft
          </Button>
          <Button
            aria-current={view === "rankings" ? "page" : undefined}
            aria-label="Ranking sources"
            onClick={() => setView("rankings")}
            disabled={!activeLeague}
          >
            Sources
          </Button>
          <Button
            aria-current={view === "leagues" ? "page" : undefined}
            aria-label="Manage leagues"
            onClick={() => setView("leagues")}
          >
            Leagues
          </Button>
          <Button
            aria-current={view === "tools" ? "page" : undefined}
            onClick={() => setView("tools")}
            disabled={!activeLeague}
          >
            Draft tools
          </Button>
          {!onboardingDone && !onboardingOpen ? (
            <Button onClick={openOnboarding}>{hasOnboardingDraft ? "Resume setup" : "Setup guide"}</Button>
          ) : null}
        </nav>
      </header>

      {error ? <StatusMessage tone="error">Unable to load leagues. {error}</StatusMessage> : null}
      {notice ? (
        <StatusMessage tone="success" ref={noticeRef} tabIndex={-1}>
          {notice}
        </StatusMessage>
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
        <DraftWorkspace
          key={activeLeagueId}
          leagueId={activeLeagueId}
          autoFocusHeading={!notice}
          mode={view === "tools" ? "tools" : "board"}
        />
      )}
    </>
  );
}
