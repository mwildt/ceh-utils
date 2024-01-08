package training

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/mwildt/ceh-utils/pkg/events"
	"github.com/mwildt/ceh-utils/pkg/utils"
	"sort"
	"time"
)

type Challenge struct {
	Id     uuid.UUID
	Answer []uuid.UUID
}

type TrainingChallenge struct {
	Id        uuid.UUID
	Answer    []uuid.UUID
	Level     int
	Timestamp time.Time
	Done      bool
	Count     int
}

func TrainingChallengeIdEquals(id uuid.UUID) utils.Predicate[*TrainingChallenge] {
	return func(q *TrainingChallenge) bool {
		return q.Id == id
	}
}

type ChallengeProvider func(excludeIds []uuid.UUID) (Challenge, error)

func Initial() utils.Predicate[*TrainingChallenge] {
	return func(q *TrainingChallenge) bool {
		return q.Level == 0 && !q.Done
	}
}

func Proceeding() utils.Predicate[*TrainingChallenge] {
	return func(q *TrainingChallenge) bool {
		return q.Level > 0 && !q.Done
	}
}

func Done() utils.Predicate[*TrainingChallenge] {
	return func(q *TrainingChallenge) bool {
		return q.Done
	}
}

func notPending(cutoff time.Time) utils.Predicate[*TrainingChallenge] {
	return func(q *TrainingChallenge) bool {
		return q.Timestamp.Before(cutoff) && !q.Done
	}
}

func (tc *TrainingChallenge) reset() {

	tc.Count = tc.Count + 1
	tc.Level = 0
	tc.Timestamp = time.Now().Add(time.Minute * 10)
}

func (tc *TrainingChallenge) proceed() {
	tc.Count = tc.Count + 1
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
	return &TrainingChallenge{challenge.Id, challenge.Answer, 0, time.Now(), false, 0}
}

func getChallengeId(c *TrainingChallenge) uuid.UUID {
	return c.Id
}

type event struct {
	Type  string
	event interface{}
}

func createdEvent(id uuid.UUID) event {
	return event{"training.created", CreatedEvent{id}}
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
	Stats                  *Stats
	Challenges             []*TrainingChallenge
	logger                 utils.Logger
}

func CreateTraining(nextChallenge ChallengeProvider) (training *Training, err error) {
	challenge, err := nextChallenge(make([]uuid.UUID, 0))
	if err != nil {
		return training, err
	}
	id := uuid.New()
	return (&Training{
		Id:                     id,
		CurrentChallenge:       createTrainingChallenge(challenge),
		currentChallengeFailed: false,
		Updated:                time.Now(),
		Created:                time.Now(),
		Challenges:             make([]*TrainingChallenge, 0),
		Stats: &Stats{
			totalChallenges:          1,
			passedChallenges:         0,
			failedChallenges:         0,
			currentChallengeAttempts: 0,
		},
	}).init(createdEvent(id)), nil
}

func (training *Training) Next(answerIds []uuid.UUID, nextChallenge ChallengeProvider) (success bool, err error) {

	success = utils.MutualContainment(training.CurrentChallenge.Answer, answerIds)

	training.events = append(training.events, event{"training.updated", UpdatedEvent{
		TrainingId:  training.Id,
		ChallengeId: training.CurrentChallenge.Id,
		AnswerIds:   answerIds,
		Passed:      success,
	}})

	if success {
		training.Stats.pass()

		if training.currentChallengeFailed {
			training.logger.Info("reset Challenge {id: %s, level: %d}", training.CurrentChallenge.Id, training.CurrentChallenge.Level)
			training.CurrentChallenge.reset()
		} else {
			training.logger.Info("proceed Challenge {id: %s, level: %d}", training.CurrentChallenge.Id, training.CurrentChallenge.Level)
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
			training.Challenges = append(training.Challenges, trainingChallenge)
			training.setCurrentChallenge(trainingChallenge)
		}
		training.Updated = time.Now()
		return success, err
	} else {
		training.currentChallengeFailed = true
		training.Stats.fail()
		return success, nil
	}
}

func (training *Training) getExcludeIds() []uuid.UUID {
	return utils.Map(training.Challenges, getChallengeId)
}

func (training *Training) findRetryCandidate() (candidate *TrainingChallenge, found bool) {
	return filterCandidates(training.Challenges)
}

func (training *Training) setCurrentChallenge(candidate *TrainingChallenge) {
	training.CurrentChallenge = candidate
	training.currentChallengeFailed = false
}

func (training *Training) init(events ...event) *Training {
	training.logger = utils.NewStdLogger(fmt.Sprintf("training-%s", training.Id.String()))
	training.events = append([]event{}, events...)
	return training
}

func (training *Training) emitEvents() {
	for _, event := range training.events {
		_ = events.Emit(event.Type, event.event)
	}
	training.events = training.events[:0]
}

func (training *Training) updateChallengeAnswer(challengeId uuid.UUID, answerId []uuid.UUID) {
	training.logger.Info("updateChallengeAnswer with id challenge Id %s to %s", challengeId, answerId)
	if training.CurrentChallenge.Id == challengeId {
		training.CurrentChallenge.Answer = answerId
	}
	for _, tq := range training.Challenges {
		if tq.Id == challengeId {
			tq.Answer = answerId
			tq.reset()
		}
	}
}

func (training *Training) GetChallengeCount(predicate utils.Predicate[*TrainingChallenge]) int {
	return utils.Count(training.Challenges, predicate)
}

func ContainsChallenge(challengeId uuid.UUID) utils.Predicate[*Training] {
	return func(q *Training) bool {
		return q.CurrentChallenge.Id == challengeId || utils.AnyMatch(q.Challenges, TrainingChallengeIdEquals(challengeId))
	}
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
