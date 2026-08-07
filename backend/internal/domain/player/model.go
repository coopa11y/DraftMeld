package player

type Candidate struct {
	IdentityKey string
	LegacyKey   string
	Name        string
	Position    string
	Team        string
	Provider    string
	ProviderID  string
}

type Player struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Position string `json:"position"`
	Team     string `json:"team"`
}

type DirectoryStatus struct {
	PlayerCount     int `json:"playerCount"`
	IdentityCount   int `json:"identityCount"`
	ProviderIDCount int `json:"providerIdCount"`
}
