package sqlite

import (
	"context"
	"fmt"
	"time"
)

func (store *DraftEventStore) IdentityReviews(ctx context.Context) (map[string]string, error) {
	rows, err := store.database.QueryContext(ctx, `SELECT issue_key, resolution FROM identity_reviews`)
	if err != nil {
		return nil, fmt.Errorf("query identity reviews: %w", err)
	}
	defer rows.Close()
	reviews := make(map[string]string)
	for rows.Next() {
		var key, resolution string
		if err = rows.Scan(&key, &resolution); err != nil {
			return nil, fmt.Errorf("scan identity review: %w", err)
		}
		reviews[key] = resolution
	}
	return reviews, rows.Err()
}

func (store *DraftEventStore) SaveIdentityReview(ctx context.Context, issueKey, resolution string) error {
	_, err := store.database.ExecContext(ctx, `INSERT INTO identity_reviews (issue_key, resolution, reviewed_at) VALUES (?, ?, ?)
ON CONFLICT(issue_key) DO UPDATE SET resolution=excluded.resolution, reviewed_at=excluded.reviewed_at`, issueKey, resolution, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("save identity review: %w", err)
	}
	return nil
}
