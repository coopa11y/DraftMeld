package api

import (
	"net/http"
	"strconv"

	"github.com/coopa11y/DraftMeld/backend/internal/application"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
)

type draftTradeRequest struct {
	LeagueID             string              `json:"leagueId"`
	TeamOneNumber        int                 `json:"teamOneNumber"`
	TeamTwoNumber        int                 `json:"teamTwoNumber"`
	TeamOneReceives      []int               `json:"teamOneReceives"`
	TeamTwoReceives      []int               `json:"teamTwoReceives"`
	TeamOneFuturePicks   []draft.FuturePick  `json:"teamOneFuturePicks"`
	TeamTwoFuturePicks   []draft.FuturePick  `json:"teamTwoFuturePicks"`
	TeamOneAuctionBudget float64             `json:"teamOneAuctionBudget"`
	TeamTwoAuctionBudget float64             `json:"teamTwoAuctionBudget"`
	TeamOnePlayers       []string            `json:"teamOnePlayers"`
	TeamTwoPlayers       []string            `json:"teamTwoPlayers"`
	TeamOneBudgets       []draft.BudgetAsset `json:"teamOneBudgets"`
	TeamTwoBudgets       []draft.BudgetAsset `json:"teamTwoBudgets"`
}

type tradeConditionRequest struct {
	LeagueID           string `json:"leagueId"`
	Season             int    `json:"season"`
	Round              int    `json:"round"`
	OriginalTeamNumber int    `json:"originalTeamNumber"`
	Status             string `json:"status"`
}

func registerDraftTradeRoutes(mux *http.ServeMux, drafts *application.DraftService) {
	mux.HandleFunc("POST /api/v1/draft/trades", func(response http.ResponseWriter, request *http.Request) {
		input, ok := decodeJSON[draftTradeRequest](response, request, "The draft asset trade was not valid.")
		if !ok || input.LeagueID == "" {
			if ok {
				writeError(response, http.StatusBadRequest, "The draft asset trade was not valid.")
			}
			return
		}
		snapshot, err := drafts.CreateDraftTrade(request.Context(), input.LeagueID, input.TeamOneNumber, input.TeamTwoNumber,
			input.TeamOneReceives, input.TeamTwoReceives, input.TeamOneFuturePicks, input.TeamTwoFuturePicks,
			input.TeamOnePlayers, input.TeamTwoPlayers, input.TeamOneBudgets, input.TeamTwoBudgets,
			input.TeamOneAuctionBudget, input.TeamTwoAuctionBudget)
		if writeServiceError(response, err, http.StatusBadRequest, "",
			serviceError{application.ErrLeagueNotFound, http.StatusNotFound, "That league was not found."}) {
			return
		}
		writeJSON(response, http.StatusCreated, snapshot)
	})

	mux.HandleFunc("PUT /api/v1/draft/trades/{tradeId}/condition", func(response http.ResponseWriter, request *http.Request) {
		tradeID, parseErr := strconv.ParseInt(request.PathValue("tradeId"), 10, 64)
		input, ok := decodeJSON[tradeConditionRequest](response, request, "A valid trade condition is required.")
		if !ok {
			return
		}
		if parseErr != nil || tradeID < 1 || input.LeagueID == "" {
			writeError(response, http.StatusBadRequest, "A valid trade condition is required.")
			return
		}
		snapshot, err := drafts.ResolveTradeCondition(request.Context(), input.LeagueID, tradeID, input.Season, input.Round, input.OriginalTeamNumber, input.Status)
		if err != nil {
			writeError(response, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(response, http.StatusOK, snapshot)
	})

	mux.HandleFunc("DELETE /api/v1/draft/trades/{tradeId}", func(response http.ResponseWriter, request *http.Request) {
		leagueID := request.URL.Query().Get("leagueId")
		tradeID, err := strconv.ParseInt(request.PathValue("tradeId"), 10, 64)
		if leagueID == "" || err != nil || tradeID < 1 {
			writeError(response, http.StatusBadRequest, "A league ID and trade ID are required.")
			return
		}
		snapshot, err := drafts.DeletePickTrade(request.Context(), leagueID, tradeID)
		if writeServiceError(response, err, http.StatusBadRequest, "",
			serviceError{application.ErrLeagueNotFound, http.StatusNotFound, "That league was not found."}) {
			return
		}
		writeJSON(response, http.StatusOK, snapshot)
	})
}
