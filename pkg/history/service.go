package history

import (
	"context"
	"github.com/mwildt/ceh-utils/pkg/events"
	"github.com/mwildt/ceh-utils/pkg/training"
	"github.com/mwildt/ceh-utils/pkg/utils"
)

func Subscribe(repository Repository) error {

	logger := utils.NewStdLogger("history.service")

	err := events.Subscribe("training.created", func(event training.CreatedEvent) error {
		logger.Info("handle event training.created for id %s", event.TrainingId)
		history := CreateHistory(event.TrainingId)
		_, err := repository.Save(context.TODO(), history)
		return err
	})

	if err != nil {
		return err
	}
	logger.Info("successfully registered to training.created")

	err = events.Subscribe("training.updated", func(event training.UpdatedEvent) error {
		logger.Info("handle event training.updated for id %s", event.TrainingId)

		history, found := repository.FindFirst(context.TODO(), IdEquals(event.TrainingId))

		if !found {
			logger.Error("unable to find history with id %s - try to create new ", event.TrainingId)
			history = CreateHistory(event.TrainingId)
			_, err := repository.Save(context.TODO(), history)
			return err
		}

		history.AddAnswer(event.AnswerIds)
		if event.Passed {
			history.Finalize(event.ChallengeId, event.AnswerIds)
		}
		_, err := repository.Save(context.TODO(), history)
		return err
	})
	if err != nil {
		return err
	}
	logger.Info("successfully registered to training.updaed")
	return nil
}
