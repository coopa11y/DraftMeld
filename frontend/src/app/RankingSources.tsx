import type { League } from "../shared/api/types";
import { RankingSourcesView } from "./RankingSourcesView";
import { useRankingSources } from "./useRankingSources";

interface RankingSourcesProps {
  league: League;
  onLeagueUpdated: (league: League) => void;
}

export function RankingSources({ league, onLeagueUpdated }: RankingSourcesProps) {
  return <RankingSourcesView controller={useRankingSources(league, onLeagueUpdated)} />;
}
