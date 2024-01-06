package training

import (
	"context"
	"github.com/mwildt/ceh-utils/pkg/events"
	"github.com/mwildt/ceh-utils/pkg/questions"
	"github.com/mwildt/ceh-utils/pkg/utils"
)

func Subscribe(repository Repository) (err error) {

	logger := utils.NewStdLogger("trainings.service")
	eventType := "question.updated"

	err = events.Subscribe(eventType, func(event questions.UpdatedEvent) error {
		logger.Info("handle event %s for id %s", eventType, event.QuestionId)

		trainings, err := repository.FindAllBy(context.Background(), ContainsChallenge(event.QuestionId))
		if err != nil {
			logger.Error("unable to find trainings to update for question Id %s", event.QuestionId)
			return err
		}
		for _, training := range trainings {
			training.updateChallengeAnswer(event.QuestionId, event.AnswerIds)
			_, err = repository.Save(context.Background(), training)
			return err
		}
		return err
	})

	if err != nil {
		return err
	}
	logger.Info("successfully registered to %s", eventType)
	return nil
}
