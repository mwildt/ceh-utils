package questions

import (
	"github.com/google/uuid"
	"github.com/mwildt/ceh-utils/pkg/utils"
	"github.com/mwildt/go-http/routing"
	"net/http"
)

type Controller struct {
	repo *FileLogRepository
}

func NewRestController(repo *FileLogRepository) *Controller {
	return &Controller{
		repo: repo,
	}
}

func (controller *Controller) Routing(router routing.Routing) {
	router.HandleFunc(routing.Get("/api/questions/"), controller.GetAll)
	router.HandleFunc(routing.Get("/api/questions/{questionId}"), controller.GetById)
}

func (controller *Controller) GetById(writer http.ResponseWriter, request *http.Request) {
	if id, exists := routing.GetParameter(request.Context(), "questionId"); !exists {
		utils.BadRequest(writer, request)
	} else if qId, err := uuid.Parse(id); err != nil {
		utils.BadRequest(writer, request)
	} else if question, exists := controller.repo.FindFirst(IdEquals(qId)); !exists {
		utils.NotFound(writer, request)
	} else {
		utils.OkJson(writer, request, mapToResponse(question))
	}
}

func (controller *Controller) GetAll(writer http.ResponseWriter, request *http.Request) {
	questions, err := controller.repo.FindAll(True())
	if err != nil {
		utils.InternalServerError(writer, request)
	} else {
		utils.OkJson(writer, request, utils.Map(questions, mapToResponse))
	}
}

type answerResponse struct {
	Id   uuid.UUID `json:"id"`
	Text string    `json:"text"`
}

type response struct {
	Id      uuid.UUID        `json:"id"`
	Text    string           `json:"text"`
	Choices []answerResponse `json:"choices"`
}

func mapToResponse(question Question) response {
	return response{
		Id:   question.Id,
		Text: question.Question,
		Choices: utils.Map(question.Options, func(opt Option) answerResponse {
			return answerResponse{opt.Id, opt.Option}
		}),
	}
}
