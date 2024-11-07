package training

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/mwildt/go-http/httputils"
	"github.com/mwildt/go-http/routing"
	"github.com/ohrenpiraten/go-collections/collections"
	"github.com/ohrenpiraten/go-collections/predicates"
	"net/http"
	"time"
)

type Controller struct {
	repo              Repository
	challengeProvider ChallengeProvider
}

func NewRestController(repo Repository, challengeProvider ChallengeProvider) *Controller {
	return &Controller{
		repo:              repo,
		challengeProvider: challengeProvider,
	}
}

func (controller *Controller) Routing(router routing.Routing) {
	router.HandleFunc(routing.Post("/api/trainings/"), controller.Post)
	router.HandleFunc(routing.Get("/api/trainings/"), controller.GetAll)
	router.HandleFunc(routing.Patch("/api/trainings/{trainingId}"), controller.PatchById)
	router.HandleFunc(routing.Get("/api/trainings/{trainingId}"), controller.GetById)
	router.HandleFunc(routing.Get("/api/trainings/{trainingId}/challenges"), controller.GetChallengesById)
}

func (controller *Controller) Post(writer http.ResponseWriter, request *http.Request) {
	if training, err := CreateTraining(controller.challengeProvider); err != nil {
		httputils.InternalServerError(writer, request)
	} else if training, err := controller.repo.Save(request.Context(), training); err != nil {
		httputils.InternalServerError(writer, request)
	} else {
		httputils.CreatedJson(writer, request, createResponseDTO{
			Id: training.Id,
		})
	}
}

func (controller *Controller) GetAll(writer http.ResponseWriter, request *http.Request) {
	trainings, err := controller.repo.FindAllBy(request.Context(), predicates.True[*Training]())
	if err != nil {
		httputils.InternalServerError(writer, request)
	} else {
		httputils.CreatedJson(writer, request, collections.Map(trainings, mapGetTrainingDTO))
	}
}

func mapGetTrainingDTO(t *Training) getTrainigDTO {
	return getTrainigDTO{
		Id:                     t.Id,
		Challenge:              t.CurrentChallenge.Id,
		CurrentChallengeFailed: t.currentChallengeFailed,
		CurrentLevel:           t.CurrentChallenge.Level,
		CurrentCount:           t.CurrentChallenge.Count,
		Updated:                t.Updated.Format(time.RFC3339),
		Created:                t.Created.Format(time.RFC3339),
		ChallengeStats: challengeStatsDTO{
			len(t.Challenges),
			t.GetChallengeCount(Initial()),
			t.GetChallengeCount(Proceeding()),
			t.GetChallengeCount(Done()),
		},
		Stats: statsDTO{
			t.Stats.totalChallenges,
			t.Stats.passedChallenges,
			t.Stats.failedChallenges,
			t.Stats.currentChallengeAttempts,
		},
	}
}

func (controller *Controller) PatchById(w http.ResponseWriter, r *http.Request) {
	type responseDTO struct {
		getTrainigDTO
		Success bool `json:"success"`
	}

	var requestDTO struct {
		Answer []uuid.UUID `json:"answer"`
	}

	if trainingId, exists := routing.GetParameter(r.Context(), "trainingId"); !exists {
		httputils.BadRequest(w, r)
	} else if trainingUuid, err := uuid.Parse(trainingId); err != nil {
		httputils.BadRequest(w, r)
	} else if training, exists := controller.repo.FindFirst(r.Context(), IdEquals(trainingUuid)); !exists {
		httputils.NotFound(w, r)
	} else if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		httputils.BadRequest(w, r)
	} else if success, err := training.Next(requestDTO.Answer, controller.challengeProvider); err != nil {
		httputils.BadRequest(w, r)
	} else if training, err = controller.repo.Save(r.Context(), training); err != nil {
		httputils.InternalServerError(w, r)
	} else {
		httputils.OkJson(w, r, responseDTO{mapGetTrainingDTO(training), success})
	}
}

func (controller *Controller) GetById(w http.ResponseWriter, r *http.Request) {
	if trainingId, exists := routing.GetParameter(r.Context(), "trainingId"); !exists {
		httputils.BadRequest(w, r)
	} else if trainingUuid, err := uuid.Parse(trainingId); err != nil {
		httputils.BadRequest(w, r)
	} else if training, exists := controller.repo.FindFirst(r.Context(), IdEquals(trainingUuid)); !exists {
		httputils.NotFound(w, r)
	} else {
		httputils.OkJson(w, r, mapGetTrainingDTO(training))
	}
}

func (controller *Controller) GetChallengesById(w http.ResponseWriter, r *http.Request) {

	type challengeDto struct {
		Id    uuid.UUID `json:"id"`
		Level int       `json:"level"`
		Count int       `json:"count"`
		Done  bool      `json:"done"`
	}

	mapChallengeDTO := func(c *TrainingChallenge) challengeDto {
		return challengeDto{Id: c.Id, Level: c.Level, Count: c.Count, Done: c.Done}
	}

	type responseDTO struct {
		Id         uuid.UUID      `json:"id"`
		Challenges []challengeDto `json:"challenges"`
	}

	if trainingId, exists := routing.GetParameter(r.Context(), "trainingId"); !exists {
		httputils.BadRequest(w, r)
	} else if trainingUuid, err := uuid.Parse(trainingId); err != nil {
		httputils.BadRequest(w, r)
	} else if training, exists := controller.repo.FindFirst(r.Context(), IdEquals(trainingUuid)); !exists {
		httputils.NotFound(w, r)
	} else {
		httputils.OkJson(w, r, responseDTO{
			Id:         training.Id,
			Challenges: collections.Map(training.Challenges, mapChallengeDTO),
		})
	}
}

type createResponseDTO struct {
	Id uuid.UUID `json:"id"`
}

type statsDTO struct {
	TotalChallenges          int `json:"total"`
	PassedChallenged         int `json:"passed"`
	FailedChallenges         int `json:"failed"`
	CurrentChallengeAttempts int `json:"currentAttempts"`
}

type challengeStatsDTO struct {
	Total      int `json:"total"`
	Initial    int `json:"initial"`
	Proceeding int `json:"proceeding"`
	Done       int `json:"done"`
}

type getTrainigDTO struct {
	Id                     uuid.UUID         `json:"id"`
	Challenge              uuid.UUID         `json:"challenge"`
	CurrentChallengeFailed bool              `json:"currentChallengeFailed"`
	CurrentLevel           int               `json:"currentLevel"`
	CurrentCount           int               `json:"currentCount"`
	Updated                string            `json:"updated"`
	Created                string            `json:"created"`
	Stats                  statsDTO          `json:"stats"`
	ChallengeStats         challengeStatsDTO `json:"challengeStats"`
}
