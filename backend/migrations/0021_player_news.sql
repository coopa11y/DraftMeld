CREATE TABLE player_news_sources (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN ('rss', 'sleeper')),
    source_url TEXT NOT NULL,
    attribution TEXT NOT NULL DEFAULT '',
    enabled INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0, 1)),
    built_in INTEGER NOT NULL DEFAULT 0 CHECK (built_in IN (0, 1)),
    refresh_minutes INTEGER NOT NULL CHECK (refresh_minutes BETWEEN 5 AND 1440),
    last_refreshed_at TEXT,
    last_error TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE player_news_events (
    id TEXT PRIMARY KEY,
    source_id TEXT NOT NULL REFERENCES player_news_sources(id) ON DELETE CASCADE,
    category TEXT NOT NULL,
    title TEXT NOT NULL,
    summary TEXT NOT NULL DEFAULT '',
    article_url TEXT NOT NULL,
    published_at TEXT NOT NULL,
    observed_at TEXT NOT NULL
);

CREATE INDEX player_news_events_published_idx ON player_news_events(published_at DESC);

CREATE TABLE player_news_event_players (
    event_id TEXT NOT NULL REFERENCES player_news_events(id) ON DELETE CASCADE,
    player_id TEXT NOT NULL REFERENCES canonical_players(id) ON DELETE CASCADE,
    PRIMARY KEY (event_id, player_id)
);

CREATE INDEX player_news_event_players_player_idx ON player_news_event_players(player_id, event_id);

CREATE TABLE player_availability (
    player_id TEXT PRIMARY KEY REFERENCES canonical_players(id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    injury TEXT NOT NULL DEFAULT '',
    practice_participation TEXT NOT NULL DEFAULT '',
    source_id TEXT NOT NULL REFERENCES player_news_sources(id),
    updated_at TEXT NOT NULL
);

INSERT INTO player_news_sources (
    id, name, kind, source_url, attribution, enabled, built_in,
    refresh_minutes, created_at, updated_at
) VALUES
    ('sleeper-player-status', 'Sleeper player status', 'sleeper',
     'https://api.sleeper.app/v1/players/nfl', 'Sleeper', 1, 1, 1440,
     strftime('%Y-%m-%dT%H:%M:%fZ', 'now'), strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    ('espn-nfl-news', 'ESPN NFL news', 'rss',
     'https://www.espn.com/espn/rss/nfl/news', 'ESPN', 1, 1, 10,
     strftime('%Y-%m-%dT%H:%M:%fZ', 'now'), strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    ('pff-news', 'PFF news', 'rss',
     'https://www.pff.com/feed', 'PFF', 1, 1, 15,
     strftime('%Y-%m-%dT%H:%M:%fZ', 'now'), strftime('%Y-%m-%dT%H:%M:%fZ', 'now'));
