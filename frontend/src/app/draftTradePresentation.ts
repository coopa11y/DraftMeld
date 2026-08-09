import type {
  BudgetAsset,
  DraftPickSlot,
  DraftPickTrade,
  DraftSnapshot,
  FutureDraftPick,
  Player,
} from "../shared/api/types";

export function selectedDraftSlots(selected: string[], indexed: Map<string, DraftPickSlot>) {
  return selected.map((key) => indexed.get(key)).filter((pick): pick is DraftPickSlot => Boolean(pick));
}

export function currentDraftPicks(slots: DraftPickSlot[]) {
  return slots.filter((pick) => pick.overallNumber > 0).map((pick) => pick.overallNumber);
}

export function futureDraftPicks(slots: DraftPickSlot[], condition: string): FutureDraftPick[] {
  return slots
    .filter((pick) => pick.overallNumber === 0)
    .map((pick) => ({
      season: pick.season,
      round: pick.round,
      originalTeamNumber: pick.originalTeamNumber,
      originalTeamName: pick.originalTeamName,
      condition: condition.trim(),
      conditionStatus: "",
    }));
}

export function updateTradeBudget(budgets: BudgetAsset[], kind: BudgetAsset["kind"], season: number, amount: number) {
  const remaining = budgets.filter((asset) => asset.kind !== kind || asset.season !== season);
  return amount > 0 ? [...remaining, { kind, season, amount }] : remaining;
}

export function selectedAssetLabels(
  slots: DraftPickSlot[],
  players: string[],
  budgets: BudgetAsset[],
  playersByID: Map<string, Player>,
  format: DraftSnapshot["leagueFormat"],
  condition: string,
) {
  return [
    ...slots.map((pick) => {
      const label = draftPickLabel(pick, format);
      return pick.overallNumber === 0 && condition.trim() ? `${label} (condition: ${condition.trim()})` : label;
    }),
    ...players.map((id) => playersByID.get(id)?.name ?? `Player ${id}`),
    ...budgets.map(budgetAssetLabel),
  ];
}

export function storedTradeAssetLabels(
  trade: DraftPickTrade,
  first: boolean,
  currentSlots: Map<string, DraftPickSlot>,
  playersByID: Map<string, Player>,
  format: DraftSnapshot["leagueFormat"],
) {
  const current = first ? trade.teamOneReceives : trade.teamTwoReceives;
  const future = first ? trade.teamOneFuturePicks : trade.teamTwoFuturePicks;
  const players = first ? trade.teamOnePlayers : trade.teamTwoPlayers;
  const budgets = first ? trade.teamOneBudgets : trade.teamTwoBudgets;
  const labels = current.map((number) => {
    const slot = currentSlots.get(`${trade.season}:${number}`);
    return slot ? draftPickLabel(slot, format) : `${trade.season} overall pick ${number}`;
  });
  labels.push(...future.map(futureDraftPickLabel));
  labels.push(...players.map((id) => playersByID.get(id)?.name ?? `Player ${id}`));
  labels.push(...budgets.map(budgetAssetLabel));
  if (!budgets.length) {
    if (first && trade.teamOneAuctionBudget > 0)
      labels.push(`$${trade.teamOneAuctionBudget.toFixed(0)} ${trade.season} auction budget`);
    if (!first && trade.teamTwoAuctionBudget > 0)
      labels.push(`$${trade.teamTwoAuctionBudget.toFixed(0)} ${trade.season} auction budget`);
  }
  return labels;
}

export function pendingTradeConditions(trade: DraftPickTrade) {
  return [...trade.teamOneFuturePicks, ...trade.teamTwoFuturePicks].filter(
    (pick) => pick.conditionStatus === "pending",
  );
}

export function futureDraftPickLabel(pick: FutureDraftPick) {
  const base = `${pick.season}, round ${pick.round}, ${pick.originalTeamName} original pick`;
  if (!pick.condition) return base;
  const status =
    pick.conditionStatus === "not-met"
      ? "condition not met; transfer void"
      : pick.conditionStatus === "met"
        ? "condition met"
        : "condition pending";
  return `${base} (${status}: ${pick.condition})`;
}

export function budgetAssetLabel(asset: BudgetAsset) {
  return `$${asset.amount.toFixed(0)} ${asset.season} ${asset.kind === "faab" ? "FAAB" : "auction budget"}`;
}

export function draftTeamName(snapshot: DraftSnapshot, teamNumber: number) {
  return snapshot.teams.find((team) => team.number === teamNumber)?.name ?? `Team ${teamNumber}`;
}

export function draftPickKey(pick: DraftPickSlot) {
  return pick.overallNumber > 0
    ? `${pick.season}:overall:${pick.overallNumber}`
    : `${pick.season}:round:${pick.round}:team:${pick.originalTeamNumber}`;
}

export function futureDraftPickKey(pick: FutureDraftPick) {
  return `${pick.season}:round:${pick.round}:team:${pick.originalTeamNumber}`;
}

export function draftPickGroupLabel(pick: DraftPickSlot, format: DraftSnapshot["leagueFormat"]) {
  return format === "dynasty" ? `${pick.season} round ${pick.round}` : `Round ${pick.round}`;
}

export function draftPickLabel(pick: DraftPickSlot, format: DraftSnapshot["leagueFormat"]) {
  if (pick.overallNumber === 0) return `${pick.season}, round ${pick.round}, ${pick.originalTeamName} original pick`;
  const current = `Round ${pick.round}, pick ${pick.pickInRound}, overall ${pick.overallNumber}`;
  return format === "dynasty" ? `${pick.season}, ${current.toLowerCase()}` : current;
}

export function draftTradeHelp(snapshot: DraftSnapshot) {
  if (snapshot.leagueFormat === "redraft")
    return snapshot.draftType === "auction"
      ? "This redraft league shows only auction budget enabled by its rules; future assets stay out of the way."
      : `This redraft league shows only unused ${snapshot.season} picks. Future seasons stay out of the way.`;
  return `This dynasty league tracks players, unused ${snapshot.season} assets, future rookie picks, and enabled budgets across seasons.`;
}
