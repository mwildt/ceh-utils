package questions

import (
	"encoding/base64"
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
	mediaPath := "config/ceh-12-cehtest.org/media"
	router.Handle(routing.Get("/api/media/**"), http.StripPrefix("/api/media", http.FileServer(http.Dir(mediaPath))))

	router.HandleFunc(routing.Get("/api/questions/"), controller.GetAll)
	router.HandleFunc(routing.Get("/api/questions/{questionId}"), controller.GetById)
	router.HandleFunc(routing.Patch("/api/questions/{questionId}").Filter(apiSecured()), controller.PatchById)

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
		Text    string           `json:"text"`
		Choices []answerResponse `json:"choices"`
		Answer  []uuid.UUID      `json:"answer"`
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
		}), requestDTO.Answer)

		if err != nil {
			utils.InternalServerError(writer, request)
		} else if updated, err := controller.repo.Save(updated); err != nil {
			utils.InternalServerError(writer, request)
		} else {
			utils.OkJson(writer, request, mapToResponse(updated))
		}
	}
}

func readUuid(parameterName string, request *http.Request) (id uuid.UUID, err error) {
	if strId, exists := routing.GetParameter(request.Context(), parameterName); !exists {
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
	Media   []string         `json:"media"`
}

func mapToResponse(question *Question) response {
	return response{
		Id:   question.Id,
		Text: question.Question,
		Choices: utils.Map(question.Options, func(opt Option) answerResponse {
			return answerResponse{opt.Id, opt.Option}
		}),
		Media: question.Media,
	}
}

func apiSecured() routing.Filter {

	return func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		apiToken := r.Header.Get("x-api-key")
		apiKey := utils.GetEnvOrDefault("API_KEY", "")
		if apiKey == "" {
			utils.NewStdLogger("apiSecurity").Warn("Unable to find apiKey, operation denied")
		}
		if apiKey == "" || apiToken != base64.StdEncoding.EncodeToString([]byte(apiKey)) {
			utils.StatusUnauthorized(w, r)
		} else {
			next(w, r)
		}
	}
}
