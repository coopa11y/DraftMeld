export type DraftAction = "draft" | "taken";

export interface Player {
  id: string;
  name: string;
  nflTeam: string;
  position: "QB" | "RB" | "WR" | "TE";
  byeWeek: number;
  overallRank: number;
  positionRank: number;
  adp: number;
  tier: number;
}

export interface Recommendation {
  player: Player;
  score: number;
  reasons: string[];
}

export interface Pick {
  eventId: number;
  number: number;
  action: DraftAction;
  player: Player;
  createdAt: string;
}

export interface DraftSnapshot {
  leagueId: string;
  leagueName: string;
  pickNumber: number;
  available: Player[];
  myTeam: Player[];
  history: Pick[];
  recommendations: Recommendation[];
  canUndo: boolean;
}
