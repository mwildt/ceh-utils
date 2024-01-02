package history

import (
	"github.com/google/uuid"
	"github.com/mwildt/ceh-utils/pkg/utils"
	"github.com/mwildt/go-http/routing"
	"net/http"
	"strconv"
)

type Controller struct {
	repo Repository
}

func NewRestController(repo Repository) *Controller {
	return &Controller{
		repo: repo,
	}
}

func (controller *Controller) Routing(router routing.Routing) {
	router.HandleFunc(routing.Get("/api/history/{historyId}"), controller.GetHistory)
	router.HandleFunc(routing.Get("/api/history/{historyId}/{historyIndex}"), controller.GetHistoryItem)
}

func (controller *Controller) GetHistory(w http.ResponseWriter, r *http.Request) {
	type responseDTO struct {
		Id    uuid.UUID `json:"id"`
		Total int       `json:"total"`
	}

	if idString, exists := routing.GetParameter(r.Context(), "historyId"); !exists {
		utils.BadRequest(w, r)
	} else if historyId, err := uuid.Parse(idString); err != nil {
		utils.BadRequest(w, r)
	} else if hist, exists := controller.repo.FindFirst(r.Context(), IdEquals(historyId)); !exists {
		utils.NotFound(w, r)
	} else {
		utils.OkJson(w, r, responseDTO{Id: hist.Id, Total: hist.Size()})
	}
}

func (controller *Controller) GetHistoryItem(w http.ResponseWriter, r *http.Request) {
	type responseDTO struct {
		Id             uuid.UUID   `json:"historyId"`
		ChallengeId    uuid.UUID   `json:"challengeId"`
		ChallengeIndex int         `json:"index"`
		GivenAnswers   []uuid.UUID `json:"givenAnswers"`
		SolvingAnswer  uuid.UUID   `json:"solvingAnswer"`
	}

	if idString, exists := routing.GetParameter(r.Context(), "historyId"); !exists {
		utils.BadRequest(w, r)
	} else if historyId, err := uuid.Parse(idString); err != nil {
		utils.BadRequest(w, r)
	} else if idxString, exists := routing.GetParameter(r.Context(), "historyIndex"); !exists {
		utils.BadRequest(w, r)
	} else if historyIndex, err := strconv.Atoi(idxString); err != nil {
		utils.BadRequest(w, r)
	} else if hist, exists := controller.repo.FindFirst(r.Context(), IdEquals(historyId)); !exists {
		utils.NotFound(w, r)
	} else if found, item := hist.HistoryItemAt(historyIndex); !found {
		w.WriteHeader(http.StatusNotFound)
	} else {
		utils.OkJson(w, r, responseDTO{
			Id:             hist.Id,
			ChallengeId:    item.ChallengeId,
			ChallengeIndex: historyIndex,
			GivenAnswers:   item.GivenAnswers,
			SolvingAnswer:  item.SolvingAnswer})
	}
}
