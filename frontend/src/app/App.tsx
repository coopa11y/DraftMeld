import { useEffect, useState } from "react";
import { listLeagues } from "../shared/api/leagues";
import type { League } from "../shared/api/types";
import { DraftWorkspace } from "./DraftWorkspace";
import { LeagueManager } from "./LeagueManager";

const ACTIVE_LEAGUE_KEY = "draftmeld.active-league.v1";

export function App() {
  const [leagues, setLeagues] = useState<League[]>([]);
  const [activeLeagueId, setActiveLeagueId] = useState(() => localStorage.getItem(ACTIVE_LEAGUE_KEY) ?? "demo");
  const [view, setView] = useState<"draft" | "leagues">("draft");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    listLeagues()
      .then((loaded) => {
        if (!active) return;
        setLeagues(loaded);
        const selected = loaded.some((league) => league.id === activeLeagueId) ? activeLeagueId : loaded[0]?.id;
        if (selected) selectLeague(selected);
        else setView("leagues");
      })
      .catch((reason: Error) => {
        if (active) setError(reason.message);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => { active = false; };
  }, []);

  function selectLeague(id: string) {
    setActiveLeagueId(id);
    localStorage.setItem(ACTIVE_LEAGUE_KEY, id);
  }

  function handleLeaguesChange(updated: League[], preferredId?: string) {
    setLeagues(updated);
    const nextId = preferredId ?? (updated.some((league) => league.id === activeLeagueId) ? activeLeagueId : updated[0]?.id);
    if (nextId) selectLeague(nextId);
    else setView("leagues");
  }

  if (loading) {
    return <main className="centered-status" aria-busy="true"><p role="status">Loading leagues…</p></main>;
  }

  return (
    <>
      <header className="app-header">
        <div>
          <a className="brand" href="/" aria-label="DraftMeld home">DraftMeld</a>
          <span className="version">v{__APP_VERSION__}</span>
        </div>
        <nav className="league-navigation" aria-label="League navigation">
          {leagues.length > 0 ? (
            <label>
              <span>Active league</span>
              <select value={activeLeagueId} onChange={(event) => { selectLeague(event.target.value); setView("draft"); }}>
                {leagues.map((league) => <option key={league.id} value={league.id}>{league.name}</option>)}
              </select>
            </label>
          ) : null}
          <button type="button" className="secondary-button" onClick={() => setView(view === "draft" ? "leagues" : "draft")} disabled={view === "leagues" && leagues.length === 0}>
            {view === "draft" ? "Manage leagues" : "Return to draft"}
          </button>
        </nav>
      </header>

      {error ? <div className="error-banner" role="alert">Unable to load leagues. {error}</div> : null}
      {view === "leagues" ? (
        <LeagueManager
          leagues={leagues}
          activeLeagueId={activeLeagueId}
          onLeaguesChange={handleLeaguesChange}
          onOpenDraft={(id) => { selectLeague(id); setView("draft"); }}
        />
      ) : (
        <DraftWorkspace leagueId={activeLeagueId} />
      )}
    </>
  );
}
