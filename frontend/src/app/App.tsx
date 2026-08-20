import { useEffect, useRef, useState } from "react";
import { createLeague, listLeagues } from "../shared/api/leagues";
import type { League, LeagueRules, MockDraftSession } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { StatusMessage } from "../shared/ui/StatusMessage";
import { DraftWorkspace } from "./DraftWorkspace";
import { LeagueManager } from "./LeagueManager";
import { OnboardingWizard } from "./OnboardingWizard";
import { RankingSources } from "./RankingSources";
import { PlayerNewsSources } from "./PlayerNewsSources";
import { loadOnboardingDraft, onboardingComplete } from "./onboarding";

const ACTIVE_LEAGUE_KEY = "draftmeld.active-league.v1";

export function App() {
  const [leagues, setLeagues] = useState<League[]>([]);
  const [activeLeagueId, setActiveLeagueId] = useState(() => localStorage.getItem(ACTIVE_LEAGUE_KEY) ?? "demo");
  const initialActiveLeagueId = useRef(activeLeagueId);
  const [view, setView] = useState<"draft" | "tools" | "leagues" | "league" | "rankings" | "news">("leagues");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [onboardingOpen, setOnboardingOpen] = useState(false);
  const [onboardingDone, setOnboardingDone] = useState(onboardingComplete);
  const [hasOnboardingDraft, setHasOnboardingDraft] = useState(() => loadOnboardingDraft() !== null);
  const [notice, setNotice] = useState("");
  const [mockDraft, setMockDraft] = useState<MockDraftSession | null>(null);
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
    setMockDraft(null);
    selectLeague(value);
    setView("league");
  }

  const activeLeague = leagues.find((league) => league.id === activeLeagueId);
  const leagueContextOpen = view !== "leagues" && Boolean(activeLeague);
  const draftLeagueId = mockDraft?.id ?? activeLeagueId;

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
          {leagueContextOpen ? (
            <label>
              <span>Active league</span>
              <select value={activeLeague?.id ?? ""} onChange={(event) => handleLeagueSelection(event.target.value)}>
                {leagues.map((league) => (
                  <option key={league.id} value={league.id}>
                    {league.name}
                  </option>
                ))}
              </select>
            </label>
          ) : null}
          {leagueContextOpen ? (
            <>
              <Button aria-current={view === "draft" ? "page" : undefined} onClick={() => setView("draft")}>
                {mockDraft ? "Mock draft" : "Draft"}
              </Button>
              <Button
                aria-current={view === "rankings" ? "page" : undefined}
                aria-label="Ranking sources"
                onClick={() => setView("rankings")}
              >
                Sources
              </Button>
            </>
          ) : null}
          <Button
            aria-current={view === "leagues" ? "page" : undefined}
            aria-label="Manage leagues"
            onClick={() => {
              setMockDraft(null);
              setView("leagues");
            }}
          >
            Leagues
          </Button>
          {leagueContextOpen ? (
            <Button aria-current={view === "tools" ? "page" : undefined} onClick={() => setView("tools")}>
              Draft tools
            </Button>
          ) : null}
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
      ) : view === "news" && activeLeague ? (
        <PlayerNewsSources onBack={() => setView("league")} />
      ) : view === "leagues" || view === "league" ? (
        <LeagueManager
          leagues={leagues}
          activeLeagueId={activeLeagueId}
          mode={view === "league" ? "overview" : "list"}
          onLeaguesChange={handleLeaguesChange}
          onOpenLeague={(id) => {
            setMockDraft(null);
            selectLeague(id);
            setView("league");
          }}
          onOpenDraft={(id) => {
            setMockDraft(null);
            selectLeague(id);
            setView("draft");
          }}
          onOpenSources={(id) => {
            setMockDraft(null);
            selectLeague(id);
            setView("rankings");
          }}
          onOpenNews={(id) => {
            setMockDraft(null);
            selectLeague(id);
            setView("news");
          }}
          onShowAll={() => {
            setMockDraft(null);
            setView("leagues");
          }}
          onOpenMockDraft={(session) => {
            selectLeague(session.leagueId);
            setMockDraft(session);
            setView("draft");
          }}
        />
      ) : (
        <DraftWorkspace
          key={draftLeagueId}
          leagueId={draftLeagueId}
          autoFocusHeading={!notice}
          mode={view === "tools" ? "tools" : "board"}
          onLeagueRulesChanged={(snapshot) => {
            if (mockDraft) return;
            setLeagues((current) =>
              current.map((league) =>
                league.id === snapshot.leagueId
                  ? {
                      ...league,
                      draftPosition: snapshot.draftPosition,
                      draftOrder: snapshot.draftOrder,
                      userTeamNumber: snapshot.userTeamNumber,
                    }
                  : league,
              ),
            );
          }}
        />
      )}
    </>
  );
}
