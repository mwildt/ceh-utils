package history

import (
	"context"
	"fmt"
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
		if history, found := repository.FindFirst(context.TODO(), IdEquals(event.TrainingId)); !found {
			return fmt.Errorf("unable to find history with id %s", event.TrainingId)
		} else {
			history.AddAnswer(event.AnswerId)
			if event.Passed {
				history.Finalize(event.ChallengeId, event.AnswerId)
			}
			_, err := repository.Save(context.TODO(), history)
			return err
		}
	})
	if err != nil {
		return err
	}
	logger.Info("successfully registered to training.updaed")
	return nil
}
