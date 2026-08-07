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

func (store *DraftEventStore) IdentityAliases(ctx context.Context) (map[string]string, error) {
	rows, err := store.database.QueryContext(ctx, `SELECT alias_key, canonical_key FROM identity_aliases`)
	if err != nil {
		return nil, fmt.Errorf("query identity aliases: %w", err)
	}
	defer rows.Close()
	aliases := make(map[string]string)
	for rows.Next() {
		var alias, canonical string
		if err = rows.Scan(&alias, &canonical); err != nil {
			return nil, fmt.Errorf("scan identity alias: %w", err)
		}
		aliases[alias] = canonical
	}
	return aliases, rows.Err()
}

func (store *DraftEventStore) SaveIdentityMerge(ctx context.Context, issueKey, canonicalKey string, aliases []string) error {
	tx, err := store.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin identity merge: %w", err)
	}
	defer tx.Rollback()
	for _, alias := range aliases {
		if alias == "" || alias == canonicalKey {
			continue
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO identity_aliases (alias_key, canonical_key, created_at) VALUES (?, ?, ?)
ON CONFLICT(alias_key) DO UPDATE SET canonical_key=excluded.canonical_key, created_at=excluded.created_at`, alias, canonicalKey, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
			return fmt.Errorf("save identity alias: %w", err)
		}
		if _, err = tx.ExecContext(ctx, `UPDATE player_identity_keys SET player_id = ? WHERE player_id = ?`, canonicalKey, alias); err != nil {
			return fmt.Errorf("merge player identity keys: %w", err)
		}
		if _, err = tx.ExecContext(ctx, `UPDATE player_provider_ids SET player_id = ? WHERE player_id = ?`, canonicalKey, alias); err != nil {
			return fmt.Errorf("merge provider player IDs: %w", err)
		}
		if _, err = tx.ExecContext(ctx, `UPDATE canonical_players SET merged_into = ?, updated_at = ? WHERE id = ?`, canonicalKey, time.Now().UTC().Format(time.RFC3339Nano), alias); err != nil {
			return fmt.Errorf("merge canonical player: %w", err)
		}
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO identity_reviews (issue_key, resolution, reviewed_at) VALUES (?, 'acknowledged', ?)
ON CONFLICT(issue_key) DO UPDATE SET resolution=excluded.resolution, reviewed_at=excluded.reviewed_at`, issueKey, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		return fmt.Errorf("save merged identity review: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit identity merge: %w", err)
	}
	return nil
}
