package ranking

import (
	"errors"
	"math"
	"sort"
)

const (
	MethodWeightedAverage = "weighted-average"
	MethodWeightedMedian  = "weighted-median"
	MethodTrimmedMean     = "trimmed-mean"
)

type Source struct {
	ID     string
	Role   string
	Weight float64
	Ranks  map[string]int
}

type Entry struct {
	PlayerID    string  `json:"playerId"`
	Score       float64 `json:"score"`
	SourceCount int     `json:"sourceCount"`
	Coverage    float64 `json:"coverage"`
	RankRange   int     `json:"rankRange"`
	Confidence  string  `json:"confidence"`
}

type signal struct {
	value   float64
	weight  float64
	present bool
}

// Combine normalizes lists of different depths onto the eligible player pool,
// then applies the selected robust consensus method. Missing entries from a
// primary ranking source receive a conservative score; contextual market and
// usage signals are only applied when they actually rank the player.
func Combine(sources []Source, eligible []string, method string) ([]Entry, error) {
	if method != MethodWeightedAverage && method != MethodWeightedMedian && method != MethodTrimmedMean {
		return nil, errors.New("unsupported consensus method")
	}
	if len(eligible) == 0 {
		return []Entry{}, nil
	}
	totalWeight := 0.0
	for _, source := range sources {
		if source.ID == "" || source.Weight <= 0 {
			return nil, errors.New("each ranking source requires an ID and positive weight")
		}
		totalWeight += source.Weight
	}
	poolSize := len(eligible)
	byPlayer := make(map[string][]signal, poolSize)
	for _, playerID := range eligible {
		if playerID == "" {
			return nil, errors.New("player IDs must be non-empty")
		}
		byPlayer[playerID] = make([]signal, 0, len(sources))
	}
	for _, source := range sources {
		depth := sourceDepth(source.Ranks)
		for playerID := range byPlayer {
			rank, present := source.Ranks[playerID]
			if present && rank < 1 {
				return nil, errors.New("ranks must be positive")
			}
			if !present && source.Role != "ranking" {
				continue
			}
			value := normalizeRank(rank, depth, poolSize)
			if !present {
				value = float64(poolSize) + math.Max(5, float64(poolSize)*0.05)
			}
			byPlayer[playerID] = append(byPlayer[playerID], signal{value: value, weight: source.Weight, present: present})
		}
	}

	entries := make([]Entry, 0, poolSize)
	for playerID, signals := range byPlayer {
		presentCount, presentWeight := 0, 0.0
		minimum, maximum := math.MaxFloat64, 0.0
		for _, item := range signals {
			if !item.present {
				continue
			}
			presentCount++
			presentWeight += item.weight
			minimum = math.Min(minimum, item.value)
			maximum = math.Max(maximum, item.value)
		}
		if presentCount == 0 {
			continue
		}
		coverage := presentWeight / totalWeight
		rankRange := int(math.Round(maximum - minimum))
		entries = append(entries, Entry{
			PlayerID: playerID, Score: combineSignals(signals, method), SourceCount: presentCount,
			Coverage: coverage, RankRange: rankRange, Confidence: confidence(coverage, rankRange),
		})
	}
	sort.Slice(entries, func(left, right int) bool {
		if entries[left].Score == entries[right].Score {
			return entries[left].PlayerID < entries[right].PlayerID
		}
		return entries[left].Score < entries[right].Score
	})
	return entries, nil
}

func WeightedAverage(sources []Source) ([]Entry, error) {
	players := make(map[string]bool)
	for _, source := range sources {
		for playerID := range source.Ranks {
			players[playerID] = true
		}
	}
	eligible := make([]string, 0, len(players))
	for playerID := range players {
		eligible = append(eligible, playerID)
	}
	return Combine(sources, eligible, MethodWeightedAverage)
}

func sourceDepth(ranks map[string]int) int {
	depth := 0
	for _, rank := range ranks {
		depth = max(depth, rank)
	}
	return depth
}

func normalizeRank(rank, depth, poolSize int) float64 {
	if poolSize <= 1 || depth <= 1 {
		return 1
	}
	return 1 + float64(rank-1)*float64(poolSize-1)/float64(depth-1)
}

func combineSignals(signals []signal, method string) float64 {
	ordered := append([]signal(nil), signals...)
	sort.Slice(ordered, func(left, right int) bool { return ordered[left].value < ordered[right].value })
	if method == MethodWeightedMedian {
		total := 0.0
		for _, item := range ordered {
			total += item.weight
		}
		threshold, cumulative := total/2, 0.0
		for _, item := range ordered {
			cumulative += item.weight
			if cumulative >= threshold {
				return item.value
			}
		}
	}
	if method == MethodTrimmedMean && len(ordered) >= 4 {
		ordered = ordered[1 : len(ordered)-1]
	}
	total, weights := 0.0, 0.0
	for _, item := range ordered {
		total += item.value * item.weight
		weights += item.weight
	}
	return total / weights
}

func confidence(coverage float64, rankRange int) string {
	if coverage >= 0.75 && rankRange <= 24 {
		return "high"
	}
	if coverage >= 0.5 && rankRange <= 60 {
		return "medium"
	}
	return "low"
}
