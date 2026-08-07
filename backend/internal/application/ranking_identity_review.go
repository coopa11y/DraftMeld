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
	SaveIdentityReview(context.Context, string, string) error
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

func (service *RankingService) ReviewIdentity(ctx context.Context, issueKey, resolution string) error {
	if issueKey == "" || (resolution != "confirmed-separate" && resolution != "acknowledged") {
		return errors.New("identity review requires an issue and valid resolution")
	}
	repository, ok := service.repository.(IdentityReviewRepository)
	if !ok {
		return errors.New("identity reviews are not supported")
	}
	return repository.SaveIdentityReview(ctx, issueKey, resolution)
}
