import { useEffect, useRef, useState } from "react";
import { listLeagues } from "../shared/api/leagues";
import type { League } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { StatusMessage } from "../shared/ui/StatusMessage";
import { DraftWorkspace } from "./DraftWorkspace";
import { LeagueManager } from "./LeagueManager";
import { RankingSources } from "./RankingSources";

const ACTIVE_LEAGUE_KEY = "draftmeld.active-league.v1";

export function App() {
  const [leagues, setLeagues] = useState<League[]>([]);
  const [activeLeagueId, setActiveLeagueId] = useState(() => localStorage.getItem(ACTIVE_LEAGUE_KEY) ?? "demo");
  const initialActiveLeagueId = useRef(activeLeagueId);
  const [view, setView] = useState<"draft" | "leagues" | "rankings">("draft");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    listLeagues()
      .then((loaded) => {
        if (!active) return;
        setLeagues(loaded);
        const selected = loaded.some((league) => league.id === initialActiveLeagueId.current) ? initialActiveLeagueId.current : loaded[0]?.id;
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

  function handleLeagueUpdated(updated: League) {
    setLeagues((current) => current.map((league) => league.id === updated.id ? updated : league));
  }

  const activeLeague = leagues.find((league) => league.id === activeLeagueId);

  if (loading) {
    return <main className="centered-status" aria-busy="true"><StatusMessage>Loading leagues...</StatusMessage></main>;
  }

  return (
    <>
      <header className="app-header">
        <div>
          <a className="brand" href="/" aria-label="DraftMeld home">DraftMeld</a>
          <span className="version">v{__APP_VERSION__}</span>
        </div>
        <nav className="league-navigation" aria-label="DraftMeld navigation">
          {leagues.length > 0 ? (
            <label>
              <span>Active league</span>
              <select value={activeLeagueId} onChange={(event) => { selectLeague(event.target.value); setView("draft"); }}>
                {leagues.map((league) => <option key={league.id} value={league.id}>{league.name}</option>)}
              </select>
            </label>
          ) : null}
          {view !== "leagues" ? <Button onClick={() => setView("leagues")}>Manage leagues</Button> : null}
          {view !== "rankings" ? <Button onClick={() => setView("rankings")} disabled={!activeLeague}>Ranking sources</Button> : null}
          {view !== "draft" ? <Button onClick={() => setView("draft")} disabled={leagues.length === 0}>Return to draft</Button> : null}
        </nav>
      </header>

      {error ? <StatusMessage tone="error">Unable to load leagues. {error}</StatusMessage> : null}
      {view === "rankings" && activeLeague ? <RankingSources key={activeLeague.id} league={activeLeague} onLeagueUpdated={handleLeagueUpdated} /> : view === "leagues" ? (
        <LeagueManager
          leagues={leagues}
          activeLeagueId={activeLeagueId}
          onLeaguesChange={handleLeaguesChange}
          onOpenDraft={(id) => { selectLeague(id); setView("draft"); }}
        />
      ) : (
        <DraftWorkspace key={activeLeagueId} leagueId={activeLeagueId} />
      )}
    </>
  );
}
