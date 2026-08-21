package news

import "time"

type SourceKind string

const (
	SourceRSS     SourceKind = "rss"
	SourceSleeper SourceKind = "sleeper"
)

type Source struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Kind            SourceKind `json:"kind"`
	URL             string     `json:"url"`
	Attribution     string     `json:"attribution"`
	Enabled         bool       `json:"enabled"`
	BuiltIn         bool       `json:"builtIn"`
	RefreshMinutes  int        `json:"refreshMinutes"`
	LastRefreshedAt time.Time  `json:"lastRefreshedAt,omitempty"`
	LastError       string     `json:"lastError,omitempty"`
}

type Event struct {
	ID          string    `json:"id"`
	SourceID    string    `json:"sourceId"`
	SourceName  string    `json:"sourceName"`
	PlayerIDs   []string  `json:"playerIds"`
	PlayerNames []string  `json:"playerNames"`
	Category    string    `json:"category"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary,omitempty"`
	URL         string    `json:"url"`
	PublishedAt time.Time `json:"publishedAt"`
	ObservedAt  time.Time `json:"observedAt"`
	Confidence  string    `json:"confidence"`
}

type Availability struct {
	PlayerID              string    `json:"playerId"`
	Status                string    `json:"status"`
	Injury                string    `json:"injury,omitempty"`
	PracticeParticipation string    `json:"practiceParticipation,omitempty"`
	SourceID              string    `json:"sourceId"`
	SourceName            string    `json:"sourceName"`
	UpdatedAt             time.Time `json:"updatedAt"`
}

type PlayerUpdate struct {
	PlayerID     string        `json:"playerId"`
	PlayerName   string        `json:"playerName"`
	Availability *Availability `json:"availability,omitempty"`
	Events       []Event       `json:"events"`
}

type Feed struct {
	UpdatedAt time.Time      `json:"updatedAt,omitempty"`
	Updates   []PlayerUpdate `json:"updates"`
}

type RefreshResult struct {
	Refreshed int      `json:"refreshed"`
	Skipped   int      `json:"skipped"`
	Errors    []string `json:"errors"`
}
