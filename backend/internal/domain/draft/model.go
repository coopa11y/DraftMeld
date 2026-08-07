package draft

import "time"

type Action string

const (
	ActionDraft Action = "draft"
	ActionTaken Action = "taken"
	ActionUndo  Action = "undo"
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
	PlayerID      string    `json:"playerId"`
	Action        Action    `json:"action"`
	TargetEventID *int64    `json:"targetEventId,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	Cost          float64   `json:"cost"`
}

type Pick struct {
	EventID   int64     `json:"eventId"`
	Number    int       `json:"number"`
	Action    Action    `json:"action"`
	Player    Player    `json:"player"`
	CreatedAt time.Time `json:"createdAt"`
	Cost      float64   `json:"cost"`
}

type Recommendation struct {
	Player  Player   `json:"player"`
	Score   float64  `json:"score"`
	Reasons []string `json:"reasons"`
}

type Snapshot struct {
	LeagueID          string           `json:"leagueId"`
	LeagueName        string           `json:"leagueName"`
	PickNumber        int              `json:"pickNumber"`
	Available         []Player         `json:"available"`
	MyTeam            []Player         `json:"myTeam"`
	History           []Pick           `json:"history"`
	Recommendations   []Recommendation `json:"recommendations"`
	CanUndo           bool             `json:"canUndo"`
	DataMode          string           `json:"dataMode"`
	ProjectionCount   int              `json:"projectionCount"`
	DraftType         string           `json:"draftType"`
	NextUserPick      int              `json:"nextUserPick"`
	AuctionBudget     float64          `json:"auctionBudget"`
	BudgetRemaining   float64          `json:"budgetRemaining"`
	AuctionInflation  float64          `json:"auctionInflation"`
	AuctionMinimumBid float64          `json:"auctionMinimumBid"`
	MaximumBid        float64          `json:"maximumBid"`
	IsUserTurn        bool             `json:"isUserTurn"`
}
