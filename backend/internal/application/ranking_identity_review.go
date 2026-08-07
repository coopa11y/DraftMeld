package application

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/ranking"
)

type IdentityReviewRepository interface {
	IdentityReviews(context.Context) (map[string]string, error)
	IdentityAliases(context.Context) (map[string]string, error)
	SaveIdentityReview(context.Context, string, string) error
	SaveIdentityMerge(context.Context, string, string, []string) error
}

func (service *RankingService) IdentityIssues(ctx context.Context) ([]ranking.IdentityIssue, error) {
	repository, ok := service.repository.(IdentityReviewRepository)
	if !ok {
		return []ranking.IdentityIssue{}, nil
	}
	records, err := service.repository.RankingRecords(ctx)
	if err != nil {
		return nil, err
	}
	reviews, err := repository.IdentityReviews(ctx)
	if err != nil {
		return nil, err
	}
	aliases, err := repository.IdentityAliases(ctx)
	if err != nil {
		return nil, err
	}
	groups := make(map[string]map[string]ranking.IdentityCandidate)
	for _, record := range records {
		record = canonicalizeRankingRecord(record)
		parts := strings.Fields(strings.TrimSpace(record.Name))
		if len(parts) < 2 || record.Team == "" {
			continue
		}
		lastName := normalizePlayerKey(parts[len(parts)-1])
		groupKey := lastName + "|" + record.Position + "|" + strings.ToUpper(record.Team)
		if groups[groupKey] == nil {
			groups[groupKey] = make(map[string]ranking.IdentityCandidate)
		}
		groups[groupKey][record.PlayerKey] = ranking.IdentityCandidate{PlayerKey: record.PlayerKey, Name: record.Name, Position: record.Position, Team: record.Team}
	}
	issues := make([]ranking.IdentityIssue, 0)
	for groupKey, candidates := range groups {
		if len(candidates) < 2 {
			continue
		}
		issue := ranking.IdentityIssue{IssueKey: groupKey, Reason: "Similar names share a team and position", Resolution: reviews[groupKey], Candidates: make([]ranking.IdentityCandidate, 0, len(candidates))}
		for _, candidate := range candidates {
			issue.Candidates = append(issue.Candidates, candidate)
		}
		sort.Slice(issue.Candidates, func(left, right int) bool { return issue.Candidates[left].Name < issue.Candidates[right].Name })
		canonical := ""
		merged := true
		for _, candidate := range issue.Candidates {
			resolved := resolveIdentityAlias(candidate.PlayerKey, aliases)
			if canonical == "" {
				canonical = resolved
			} else if resolved != canonical {
				merged = false
			}
		}
		if merged {
			issue.Resolution, issue.CanonicalPlayerKey = "merged", canonical
		}
		issues = append(issues, issue)
	}
	sort.Slice(issues, func(left, right int) bool {
		if issues[left].Resolution == "" && issues[right].Resolution != "" {
			return true
		}
		if issues[left].Resolution != "" && issues[right].Resolution == "" {
			return false
		}
		return issues[left].IssueKey < issues[right].IssueKey
	})
	return issues, nil
}

func (service *RankingService) ReviewIdentity(ctx context.Context, issueKey, resolution, canonicalPlayerKey string) error {
	if issueKey == "" || (resolution != "confirmed-separate" && resolution != "acknowledged" && resolution != "merged") {
		return errors.New("identity review requires an issue and valid resolution")
	}
	repository, ok := service.repository.(IdentityReviewRepository)
	if !ok {
		return errors.New("identity reviews are not supported")
	}
	if resolution != "merged" {
		return repository.SaveIdentityReview(ctx, issueKey, resolution)
	}
	issues, err := service.IdentityIssues(ctx)
	if err != nil {
		return err
	}
	for _, issue := range issues {
		if issue.IssueKey != issueKey {
			continue
		}
		if issue.Resolution == "merged" {
			return errors.New("identity issue is already merged")
		}
		aliases := make([]string, 0, len(issue.Candidates)-1)
		validCanonical := false
		for _, candidate := range issue.Candidates {
			if candidate.PlayerKey == canonicalPlayerKey {
				validCanonical = true
			} else {
				aliases = append(aliases, candidate.PlayerKey)
			}
		}
		if !validCanonical {
			return errors.New("choose a canonical player from the identity candidates")
		}
		return repository.SaveIdentityMerge(ctx, issueKey, canonicalPlayerKey, aliases)
	}
	return errors.New("identity review issue was not found")
}

func (service *RankingService) identityAliases(ctx context.Context) (map[string]string, error) {
	repository, ok := service.repository.(IdentityReviewRepository)
	if !ok {
		return map[string]string{}, nil
	}
	return repository.IdentityAliases(ctx)
}

func resolveIdentityAlias(playerKey string, aliases map[string]string) string {
	seen := make(map[string]bool)
	for aliases[playerKey] != "" && !seen[playerKey] {
		seen[playerKey] = true
		playerKey = aliases[playerKey]
	}
	return playerKey
}
