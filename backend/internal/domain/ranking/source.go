package ranking

import "time"

type SourceDefinition struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Methodology   string  `json:"methodology"`
	License       string  `json:"license"`
	ProjectURL    string  `json:"projectUrl"`
	DataURL       string  `json:"dataUrl"`
	DefaultWeight float64 `json:"defaultWeight"`
	ImportMode    string  `json:"importMode"`
	Role          string  `json:"role"`
	IsCustom      bool    `json:"isCustom"`
}

type Record struct {
	SourceID  string
	PlayerKey string
	Name      string
	Position  string
	Team      string
	Rank      int
	ADP       float64
	Tier      int
}

type SourceStatus struct {
	SourceDefinition
	RecordCount int        `json:"recordCount"`
	RefreshedAt *time.Time `json:"refreshedAt"`
	PublishedAt string     `json:"publishedAt,omitempty"`
}

type PlayerRanking struct {
	PlayerKey   string         `json:"playerKey"`
	Name        string         `json:"name"`
	Position    string         `json:"position"`
	Team        string         `json:"team"`
	Rank        int            `json:"rank"`
	Score       float64        `json:"score"`
	SourceCount int            `json:"sourceCount"`
	SourceRanks map[string]int `json:"sourceRanks"`
	Coverage    float64        `json:"coverage"`
	RankRange   int            `json:"rankRange"`
	Confidence  string         `json:"confidence"`
	Method      string         `json:"method"`
	ADP         float64        `json:"adp"`
	Tier        int            `json:"tier"`
}

type WatchlistSignal struct {
	SourceID    string `json:"sourceId"`
	SourceName  string `json:"sourceName"`
	SourceRank  int    `json:"sourceRank"`
	SpotsHigher int    `json:"spotsHigher"`
}

type WatchlistPlayer struct {
	PlayerKey     string            `json:"playerKey"`
	Name          string            `json:"name"`
	Position      string            `json:"position"`
	Team          string            `json:"team"`
	ConsensusRank *int              `json:"consensusRank"`
	Signals       []WatchlistSignal `json:"signals"`
}

type IdentityCandidate struct {
	PlayerKey string `json:"playerKey"`
	Name      string `json:"name"`
	Position  string `json:"position"`
	Team      string `json:"team"`
}

type IdentityIssue struct {
	IssueKey           string              `json:"issueKey"`
	Reason             string              `json:"reason"`
	Candidates         []IdentityCandidate `json:"candidates"`
	Resolution         string              `json:"resolution"`
	CanonicalPlayerKey string              `json:"canonicalPlayerKey,omitempty"`
}
