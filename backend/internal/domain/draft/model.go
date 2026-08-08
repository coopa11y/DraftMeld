package draft

import "time"

type Action string
type SessionStatus string

const (
	ActionDraft Action = "draft"
	ActionTaken Action = "taken"
	ActionUndo  Action = "undo"
)

const (
	SessionNotStarted SessionStatus = "not-started"
	SessionInProgress SessionStatus = "in-progress"
	SessionComplete   SessionStatus = "complete"
)

type Player struct {
	ID                   string  `json:"id"`
	Name                 string  `json:"name"`
	NFLTeam              string  `json:"nflTeam"`
	Position             string  `json:"position"`
	ByeWeek              int     `json:"byeWeek"`
	OverallRank          int     `json:"overallRank"`
	PositionRank         int     `json:"positionRank"`
	ADP                  float64 `json:"adp"`
	Tier                 int     `json:"tier"`
	ProjectedPoints      float64 `json:"projectedPoints"`
	ValueOverReplacement float64 `json:"valueOverReplacement"`
	Confidence           string  `json:"confidence"`
	RankRange            int     `json:"rankRange"`
	Preference           string  `json:"preference"`
	AuctionValue         float64 `json:"auctionValue"`
}

type Event struct {
	ID            int64     `json:"id"`
	LeagueID      string    `json:"leagueId"`
	Season        int       `json:"season"`
	PlayerID      string    `json:"playerId"`
	Action        Action    `json:"action"`
	TargetEventID *int64    `json:"targetEventId,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	Cost          float64   `json:"cost"`
	TeamNumber    int       `json:"teamNumber"`
}

type Session struct {
	LeagueID    string        `json:"leagueId"`
	Season      int           `json:"season"`
	Status      SessionStatus `json:"status"`
	StartedAt   time.Time     `json:"startedAt,omitempty"`
	UpdatedAt   time.Time     `json:"updatedAt"`
	ResetEvents []Event       `json:"resetEvents,omitempty"`
	ResetStatus SessionStatus `json:"resetStatus,omitempty"`
}

type Pick struct {
	EventID    int64     `json:"eventId"`
	Number     int       `json:"number"`
	Action     Action    `json:"action"`
	Player     Player    `json:"player"`
	CreatedAt  time.Time `json:"createdAt"`
	Cost       float64   `json:"cost"`
	TeamNumber int       `json:"teamNumber"`
	TeamName   string    `json:"teamName"`
}

type Team struct {
	Number                 int      `json:"number"`
	Name                   string   `json:"name"`
	IsUser                 bool     `json:"isUser"`
	Roster                 []Player `json:"roster"`
	AuctionBudgetRemaining float64  `json:"auctionBudgetRemaining"`
}

type PickSlot struct {
	Season             int    `json:"season"`
	OverallNumber      int    `json:"overallNumber"`
	Round              int    `json:"round"`
	PickInRound        int    `json:"pickInRound"`
	OriginalTeamNumber int    `json:"originalTeamNumber"`
	OriginalTeamName   string `json:"originalTeamName"`
	OwnerTeamNumber    int    `json:"ownerTeamNumber"`
	OwnerTeamName      string `json:"ownerTeamName"`
	IsUsed             bool   `json:"isUsed"`
}

type FuturePick struct {
	Season             int    `json:"season"`
	Round              int    `json:"round"`
	OriginalTeamNumber int    `json:"originalTeamNumber"`
	OriginalTeamName   string `json:"originalTeamName"`
	Condition          string `json:"condition"`
	ConditionStatus    string `json:"conditionStatus"`
}

type BudgetAsset struct {
	Kind   string  `json:"kind"`
	Season int     `json:"season"`
	Amount float64 `json:"amount"`
}

type BudgetBalance struct {
	TeamNumber int     `json:"teamNumber"`
	Season     int     `json:"season"`
	Kind       string  `json:"kind"`
	Remaining  float64 `json:"remaining"`
}

type PickTrade struct {
	ID                   int64         `json:"id"`
	LeagueID             string        `json:"leagueId"`
	TeamOneNumber        int           `json:"teamOneNumber"`
	TeamOneName          string        `json:"teamOneName"`
	TeamTwoNumber        int           `json:"teamTwoNumber"`
	TeamTwoName          string        `json:"teamTwoName"`
	TeamOneReceives      []int         `json:"teamOneReceives"`
	TeamTwoReceives      []int         `json:"teamTwoReceives"`
	TeamOneFuturePicks   []FuturePick  `json:"teamOneFuturePicks"`
	TeamTwoFuturePicks   []FuturePick  `json:"teamTwoFuturePicks"`
	TeamOneAuctionBudget float64       `json:"teamOneAuctionBudget"`
	TeamTwoAuctionBudget float64       `json:"teamTwoAuctionBudget"`
	TeamOnePlayers       []string      `json:"teamOnePlayers"`
	TeamTwoPlayers       []string      `json:"teamTwoPlayers"`
	TeamOneBudgets       []BudgetAsset `json:"teamOneBudgets"`
	TeamTwoBudgets       []BudgetAsset `json:"teamTwoBudgets"`
	Season               int           `json:"season"`
	CreatedAt            time.Time     `json:"createdAt"`
}

type Recommendation struct {
	Player  Player   `json:"player"`
	Score   float64  `json:"score"`
	Reasons []string `json:"reasons"`
}

type Snapshot struct {
	LeagueID            string           `json:"leagueId"`
	LeagueName          string           `json:"leagueName"`
	PickNumber          int              `json:"pickNumber"`
	Available           []Player         `json:"available"`
	MyTeam              []Player         `json:"myTeam"`
	Teams               []Team           `json:"teams"`
	History             []Pick           `json:"history"`
	Recommendations     []Recommendation `json:"recommendations"`
	CanUndo             bool             `json:"canUndo"`
	DataMode            string           `json:"dataMode"`
	ProjectionCount     int              `json:"projectionCount"`
	DraftType           string           `json:"draftType"`
	NextUserPick        int              `json:"nextUserPick"`
	AuctionBudget       float64          `json:"auctionBudget"`
	BudgetRemaining     float64          `json:"budgetRemaining"`
	AuctionInflation    float64          `json:"auctionInflation"`
	AuctionMinimumBid   float64          `json:"auctionMinimumBid"`
	MaximumBid          float64          `json:"maximumBid"`
	IsUserTurn          bool             `json:"isUserTurn"`
	TotalPicks          int              `json:"totalPicks"`
	IsComplete          bool             `json:"isComplete"`
	OnClockTeamNumber   int              `json:"onClockTeamNumber"`
	PickSlots           []PickSlot       `json:"pickSlots"`
	PickTrades          []PickTrade      `json:"pickTrades"`
	LeagueFormat        string           `json:"leagueFormat"`
	Season              int              `json:"season"`
	AuctionBudgetTrades bool             `json:"auctionBudgetTrades"`
	FAABTrades          bool             `json:"faabTrades"`
	BudgetBalances      []BudgetBalance  `json:"budgetBalances"`
	DraftOrder          []int            `json:"draftOrder"`
	UserTeamNumber      int              `json:"userTeamNumber"`
	SessionStatus       SessionStatus    `json:"sessionStatus"`
	CanReset            bool             `json:"canReset"`
	CanUndoReset        bool             `json:"canUndoReset"`
}
