package training

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/mwildt/ceh-utils/pkg/utils"
	"sort"
	"time"
)

type Challenge struct {
	Id     uuid.UUID
	Answer uuid.UUID
}

type TrainingChallenge struct {
	Id        uuid.UUID
	Answer    uuid.UUID
	Level     int
	Timestamp time.Time
	Done      bool
}

func (tc *TrainingChallenge) reset() {
	tc.Level = 0
	tc.Timestamp = time.Now().Add(time.Minute * 10)
}

func (tc *TrainingChallenge) proceed() {
	tc.Level = tc.Level + 1
	switch tc.Level {
	case 1:
		{
			tc.Timestamp = time.Now().Add(time.Minute * 10)
		}
	case 2:
		{
			tc.Timestamp = time.Now().Add(time.Hour * 6)
		}
	case 3:
		{
			tc.Timestamp = time.Now().Add(time.Hour * 24)
		}
	case 4:
		{
			tc.Timestamp = time.Now()
			tc.Done = true
		}
	}
}

func createTrainingChallenge(challenge Challenge) *TrainingChallenge {
	return &TrainingChallenge{challenge.Id, challenge.Answer, 0, time.Now(), false}
}

func getChallengeId(c *TrainingChallenge) uuid.UUID {
	return c.Id
}

type event struct {
	Type  string
	event interface{}
}

type Stats struct {
	totalChallenges          int
	passedChallenges         int
	failedChallenges         int
	currentChallengeAttempts int
}

func (stats *Stats) pass() {
	stats.totalChallenges++
	if stats.currentChallengeAttempts == 0 {
		stats.passedChallenges++
	}
	stats.currentChallengeAttempts = 0
}

func (stats *Stats) fail() {
	if stats.currentChallengeAttempts == 0 {
		stats.failedChallenges++
	}
	stats.currentChallengeAttempts++
}

type Training struct {
	Id                     uuid.UUID
	CurrentChallenge       *TrainingChallenge
	currentChallengeFailed bool
	Updated                time.Time
	Created                time.Time
	events                 []event
	stats                  *Stats
	challenges             []*TrainingChallenge
	logger                 utils.Logger
}

type ChallengeProvider func(excludeIds []uuid.UUID) (Challenge, error)

func CreateTraining(nextChallenge ChallengeProvider) (training *Training, err error) {
	challenge, err := nextChallenge(make([]uuid.UUID, 0))
	if err != nil {
		return training, err
	}
	id := uuid.New()
	return &Training{
		Id:                     id,
		CurrentChallenge:       createTrainingChallenge(challenge),
		currentChallengeFailed: false,
		Updated:                time.Now(),
		Created:                time.Now(),
		events:                 []event{{"training.created", CreatedEvent{id}}},
		stats: &Stats{
			totalChallenges:          1,
			passedChallenges:         0,
			failedChallenges:         0,
			currentChallengeAttempts: 0,
		},
		logger: utils.NewStdLogger(fmt.Sprintf("training-%s", id.String())),
	}, nil
}

func (training *Training) Next(answerId uuid.UUID, nextChallenge ChallengeProvider) (success bool, err error) {

	success = training.CurrentChallenge.Answer == answerId

	training.events = append(training.events, event{"training.updated", UpdatedEvent{
		TrainingId:  training.Id,
		ChallengeId: training.CurrentChallenge.Id,
		AnswerId:    answerId,
		Passed:      success,
	}})

	if success {
		training.stats.pass()

		if training.currentChallengeFailed {
			training.CurrentChallenge.reset()
		} else {
			training.CurrentChallenge.proceed()
		}

		if candidate, found := training.findRetryCandidate(); found {
			training.logger.Info("found retry candidate question %s %d", candidate.Id, candidate.Level)
			training.setCurrentChallenge(candidate)
		} else {
			challenge, err := nextChallenge(training.getExcludeIds())
			training.logger.Info("no retry challenge found, got new one from provider %s", challenge.Id)
			if err != nil {
				return success, err
			}
			trainingChallenge := createTrainingChallenge(challenge)
			training.challenges = append(training.challenges, trainingChallenge)
			training.setCurrentChallenge(trainingChallenge)
		}
		training.Updated = time.Now()
		return success, err
	} else {
		training.currentChallengeFailed = true
		training.stats.fail()
		return success, nil
	}
}

func (training *Training) getExcludeIds() []uuid.UUID {
	return utils.Map(training.challenges, getChallengeId)
}

func (training *Training) findRetryCandidate() (candidate *TrainingChallenge, found bool) {
	for _, c := range training.challenges {
		training.logger.Info("challenge %s level: %d: timestamp: %s", c.Id, c.Level, c.Timestamp)
	}
	return filterCandidates(training.challenges)
}

func filterCandidates(trainingChallenges []*TrainingChallenge) (candidate *TrainingChallenge, found bool) {
	// erstmal alle rausfilter, die die cutoff grente noch nicht erreich haben
	candidates := utils.Filter(trainingChallenges, notPending(time.Now()))

	// und dann muss das noch nach den passenden elementen
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Level == candidates[j].Level {
			return candidates[i].Timestamp.Before(candidates[j].Timestamp)
		} else {
			return candidates[i].Level < candidates[j].Level
		}
	})
	if len(candidates) > 0 {
		return candidates[0], true
	}
	return candidate, false
}

func notPending(cutoff time.Time) utils.Predicate[*TrainingChallenge] {
	return func(q *TrainingChallenge) bool {
		return q.Timestamp.Before(cutoff) && !q.Done
	}
}

func (training *Training) setCurrentChallenge(candidate *TrainingChallenge) {
	training.CurrentChallenge = candidate
	training.currentChallengeFailed = false
}
