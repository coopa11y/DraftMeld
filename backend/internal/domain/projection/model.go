package projection

import "time"

type Record struct {
	SourceID   string             `json:"sourceId"`
	PlayerKey  string             `json:"playerKey"`
	Name       string             `json:"name"`
	Position   string             `json:"position"`
	Team       string             `json:"team"`
	ByeWeek    int                `json:"byeWeek"`
	ADP        float64            `json:"adp"`
	Stats      map[string]float64 `json:"stats"`
	ProviderID string             `json:"-"`
}

type SourceStatus struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	RecordCount int       `json:"recordCount"`
	ImportedAt  time.Time `json:"importedAt"`
}

type LeagueValue struct {
	ProjectedPoints float64
	ADP             float64
	ByeWeek         int
}
