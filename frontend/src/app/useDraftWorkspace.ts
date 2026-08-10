import { useEffect, useRef, useState, type Dispatch, type RefObject, type SetStateAction } from "react";
import { getDraft } from "../shared/api/draft";
import { getRankingWatchlist } from "../shared/api/rankings";
import type { DraftSnapshot, WatchlistPlayer } from "../shared/api/types";
import { useViewHeadingFocus } from "../shared/hooks/useViewHeadingFocus";
import {
  createDraftWorkspaceHandlers,
  defaultSelectedTeam,
  type DraftWorkspaceHandlers,
} from "./draftWorkspaceHandlers";

export interface DraftWorkspaceController {
  announcement: string;
  boardHeading: RefObject<HTMLHeadingElement | null>;
  busy: boolean;
  error: string;
  errorAlert: RefObject<HTMLDivElement | null>;
  handlers: DraftWorkspaceHandlers;
  selectedTeamNumber: number;
  setSelectedTeamNumber: Dispatch<SetStateAction<number>>;
  snapshot: DraftSnapshot | null;
  watchlist: WatchlistPlayer[];
}

export function useDraftWorkspace(
  leagueId: string,
  autoFocusHeading = true,
  onLeagueRulesChanged?: (snapshot: DraftSnapshot) => void,
): DraftWorkspaceController {
  const [snapshot, setSnapshot] = useState<DraftSnapshot | null>(null);
  const [watchlist, setWatchlist] = useState<WatchlistPlayer[]>([]);
  const [busy, setBusy] = useState(false);
  const [announcement, setAnnouncement] = useState("Draft board loading.");
  const [error, setError] = useState("");
  const [selectedTeamNumber, setSelectedTeamNumber] = useState(0);
  const [pendingFocus, setPendingFocus] = useState<string | null>(null);
  const boardHeading = useViewHeadingFocus<HTMLHeadingElement>(snapshot !== null && autoFocusHeading);
  const errorAlert = useRef<HTMLDivElement>(null);

  useEffect(() => {
    let active = true;
    getDraft(leagueId)
      .then((data) => {
        if (!active) return;
        setSnapshot(data);
        if (data.dataMode !== "demo") {
          void getRankingWatchlist(leagueId)
            .then((loaded) => setWatchlist(Array.isArray(loaded) ? loaded : []))
            .catch(() => setWatchlist([]));
        }
        setSelectedTeamNumber(defaultSelectedTeam(data));
        setAnnouncement(`Draft board loaded. ${data.available.length} players are available.`);
      })
      .catch((reason: Error) => {
        if (active) setError(reason.message);
      });
    return () => {
      active = false;
    };
  }, [leagueId]);

  useEffect(() => {
    if (!snapshot || !pendingFocus) return;
    const requested = document.getElementById(`draft-${pendingFocus}`);
    const fallback = document.querySelector<HTMLButtonElement>("[data-player-action='primary']");
    (requested ?? fallback ?? boardHeading.current)?.focus();
    queueMicrotask(() => setPendingFocus(null));
  }, [boardHeading, pendingFocus, snapshot]);

  useEffect(() => {
    if (error) errorAlert.current?.focus();
  }, [error]);

  const handlers = createDraftWorkspaceHandlers({
    leagueId,
    snapshot,
    busy,
    onLeagueRulesChanged,
    setPendingFocus,
    setAnnouncement,
    setBusy,
    setError,
    setSelectedTeamNumber,
    setSnapshot,
  });

  return {
    announcement,
    boardHeading,
    busy,
    error,
    errorAlert,
    handlers,
    selectedTeamNumber,
    setSelectedTeamNumber,
    snapshot,
    watchlist,
  };
}
