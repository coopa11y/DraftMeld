package application

import (
	"context"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/news"
)

func (service *DraftService) overlayPlayerNews(ctx context.Context, players []draft.Player, playerByID map[string]draft.Player) {
	if service.playerNews == nil {
		return
	}
	feed, err := service.playerNews.Feed(ctx)
	if err != nil {
		return
	}
	updates := make(map[string]*news.PlayerUpdate, len(feed.Updates))
	for index := range feed.Updates {
		updates[feed.Updates[index].PlayerID] = &feed.Updates[index]
	}
	for index := range players {
		players[index].News = updates[players[index].ID]
		playerByID[players[index].ID] = players[index]
	}
}
