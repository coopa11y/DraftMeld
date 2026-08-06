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
}

type Record struct {
	SourceID  string
	PlayerKey string
	Name      string
	Position  string
	Team      string
	Rank      int
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
}
