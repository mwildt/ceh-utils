package main

import (
	"github.com/google/uuid"
	"github.com/mwildt/ceh-utils/pkg/history"
	"github.com/mwildt/ceh-utils/pkg/questions"
	"github.com/mwildt/ceh-utils/pkg/training"
	"github.com/mwildt/ceh-utils/pkg/utils"
	"github.com/mwildt/go-http/httputils"
	"github.com/mwildt/go-http/routing"
	"log"
	"net/http"
	"path"
)

func main() {

	dataPath := utils.GetEnvOrDefault("DATA_DIR", "data/")

	questionRepo, err := questions.CreateRepo(
		path.Join(dataPath, "question.data"),
		path.Join("config/ceh-12-cehtest.org", "question.data"),
		path.Join("config/custom-json", "question.data"))

	if err != nil {
		log.Fatal(err)
	}
	questionsController := questions.NewRestController(questionRepo)
	trainingRepo, err := training.CreateFileRepository(path.Join(dataPath, "trainings.data"))
	if err != nil {
		log.Fatal(err)
	}

	historyRepo, err := history.CreateRepo()
	if err != nil {
		log.Fatal(err)
	}

	if err = history.Subscribe(historyRepo); err != nil {
		log.Fatal(err)
	}
	trainingController := training.NewRestController(trainingRepo, func(excluedIds []uuid.UUID) (training.Challenge, error) {
		q, err := questionRepo.FindRandom(questions.IdNotIn(excluedIds))
		return training.Challenge{
			Id:     q.Id,
			Answer: q.AnswerIds,
		}, err
	})

	if err = training.Subscribe(trainingRepo); err != nil {
		log.Fatal(err)
	}

	baseHandler := routing.NewRouter()

	baseHandler.Route(
		routing.Filtering(requestLoggingFilter(utils.NewStdLogger("http-request-trace"))),
		questionsController.Routing,
		trainingController.Routing,
		history.NewRestController(historyRepo).Routing,
		func(router routing.Routing) {
			router.HandleFunc(routing.Path("/"), httputils.NotFound)
		},
	)

	err = http.ListenAndServe(
		utils.GetEnvOrDefault("LISTEN_ADDRESS", ":8080"),
		baseHandler)

	if err != nil {
		log.Fatal(err)
	}
}

func requestLoggingFilter(logger utils.Logger) routing.Filter {

	return func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		logger.Info("%s %s", r.Method, r.URL.String())
		next(w, r)
	}
}
