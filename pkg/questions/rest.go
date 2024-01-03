package questions

import (
	"encoding/json"
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
	router.HandleFunc(routing.Patch("/api/questions/{questionId}"), controller.PatchById)
}

func (controller *Controller) GetById(writer http.ResponseWriter, request *http.Request) {
	if questionId, err := readUuid("questionId", request); err != nil {
		utils.BadRequest(writer, request)
	} else if question, exists := controller.repo.FindFirst(IdEquals(questionId)); !exists {
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

func (controller *Controller) PatchById(writer http.ResponseWriter, request *http.Request) {
	type patchByIdRequestDTO struct {
		Text     string           `json:"text"`
		Choices  []answerResponse `json:"choices"`
		AnswerId uuid.UUID        `json:"answerId"`
	}

	if questionId, err := readUuid("questionId", request); err != nil {
		utils.BadRequest(writer, request)
	} else if requestDTO, err := readJsonPayload[patchByIdRequestDTO](request); err != nil {
		utils.BadRequest(writer, request)
	} else if question, exists := controller.repo.FindFirst(IdEquals(questionId)); !exists {
		utils.NotFound(writer, request)
	} else {
		updated, err := question.Update(requestDTO.Text, utils.Map(requestDTO.Choices, func(c answerResponse) Option {
			return Option{Option: c.Text, Id: c.Id}
		}), requestDTO.AnswerId)

		if err != nil {
			utils.InternalServerError(writer, request)
		}

		if updated, err := controller.repo.Save(updated); err != nil {
			utils.InternalServerError(writer, request)
		} else {
			utils.OkJson(writer, request, mapToResponse(updated))
		}
	}
}

func readUuid(parameterName string, request *http.Request) (id uuid.UUID, err error) {
	if strId, exists := routing.GetParameter(request.Context(), "questionId"); !exists {
		return id, err
	} else if id, err := uuid.Parse(strId); err != nil {
		return id, err
	} else {
		return id, err
	}
}

func readJsonPayload[T any](request *http.Request) (result T, err error) {
	var value T
	err = json.NewDecoder(request.Body).Decode(&value)
	return value, err
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
