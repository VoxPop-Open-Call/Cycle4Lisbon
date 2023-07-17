package achievements

import (
	"fmt"
	"math"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Store interface {
	Get(userID uuid.UUID, code string, tx *gorm.DB) (models.UserAchievement, error)
	Set(userID uuid.UUID, code string, state bool, tx *gorm.DB) (models.UserAchievement, error)
	SetCompletion(userID uuid.UUID, code string, value float64, tx *gorm.DB) (models.UserAchievement, error)
}

type Service struct {
	db    *gorm.DB
	store Store
}

// New initializes the achievements service.
func New(db *gorm.DB, store Store) (*Service, error) {
	achs := make([]models.Achievement, len(list))
	for i, a := range list {
		achs[i] = a.Achievement
		achs[i].ImageURI = fmt.Sprintf("/public/assets/achievements/%s.svg", a.Code)
	}

	err := db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(achs, 50).Error

	return &Service{db, store}, err
}

type State struct {
	Rides       uint
	Distance    float64
	Credits     float64
	Initiatives int64
}

func (s *Service) Update(userID uuid.UUID, state State) ([]models.UserAchievement, error) {
	newAchievements := make([]models.UserAchievement, 0, 5)

	err := s.db.Transaction(func(tx *gorm.DB) error {
		for _, ach := range list {
			dbAch, err := s.store.Get(userID, ach.Code, tx)
			if err != nil {
				return err
			}

			if ach.completion != nil {
				dbAch, err = s.store.SetCompletion(
					userID, ach.Code,
					ach.completion(state),
					tx,
				)
				if err != nil {
					return err
				}
			}

			if ach.trigger(state) && !dbAch.Achieved {
				newAch, err := s.store.Set(userID, ach.Code, true, tx)
				if err != nil {
					return err
				}

				newAchievements = append(newAchievements, newAch)
			}
		}

		return nil
	})

	return newAchievements, err
}

type achievement struct {
	models.Achievement
	trigger    func(s State) bool
	completion func(s State) float64
}

// list contains all achievements and their respective triggers.
var list = [...]achievement{
	// Rides
	{
		Achievement: models.Achievement{
			Code: "rides-beginner",
			Name: "Beginner",
			Desc: "You submitted your first ride!",
		},
		trigger: func(s State) bool {
			return s.Rides >= 1
		},
		completion: func(s State) float64 {
			return math.Min(float64(s.Rides), 1)
		},
	},
	{
		Achievement: models.Achievement{
			Code: "rides-traveler",
			Name: "Traveler",
			Desc: "You completed 5 rides",
		},
		trigger: func(s State) bool {
			return s.Rides >= 5
		},
		completion: func(s State) float64 {
			return math.Min(float64(s.Rides)/5, 1)
		},
	},
	{
		Achievement: models.Achievement{
			Code: "rides-pro",
			Name: "Pro",
			Desc: "You completed 100 rides",
		},
		trigger: func(s State) bool {
			return s.Rides >= 100
		},
		completion: func(s State) float64 {
			return math.Min(float64(s.Rides)/100, 1)
		},
	},

	// Distance
	{
		Achievement: models.Achievement{
			Code: "dst-training-wheels",
			Name: "Training Wheels",
			Desc: "You rode a distance of 1km",
		},
		trigger: func(s State) bool {
			return s.Distance >= 1.0
		},
		completion: func(s State) float64 {
			return math.Min(s.Distance, 1)
		},
	},
	{
		Achievement: models.Achievement{
			Code: "dst-steady-rider",
			Name: "Steady Rider",
			Desc: "You rode a distance of 50km",
		},
		trigger: func(s State) bool {
			return s.Distance >= 50.0
		},
		completion: func(s State) float64 {
			return math.Min(s.Distance/50, 1)
		},
	},
	{
		Achievement: models.Achievement{
			Code: "dst-road-champion",
			Name: "Road Champion",
			Desc: "You rode a distance of 500km",
		},
		trigger: func(s State) bool {
			return s.Distance >= 500.0
		},
		completion: func(s State) float64 {
			return math.Min(s.Distance/500, 1)
		},
	},

	// Initiatives
	{
		Achievement: models.Achievement{
			Code: "ini-good-kid",
			Name: "Good Kid",
			Desc: "You have helped your first initiative!",
		},
		trigger: func(s State) bool {
			return s.Initiatives >= 1
		},
		completion: func(s State) float64 {
			return math.Min(float64(s.Initiatives), 1)
		},
	},
	{
		Achievement: models.Achievement{
			Code: "ini-heart-of-gold",
			Name: "Heart of Gold",
			Desc: "You have helped 5 initiatives",
		},
		trigger: func(s State) bool {
			return s.Initiatives >= 5
		},
		completion: func(s State) float64 {
			return math.Min(float64(s.Initiatives)/5, 1)
		},
	},
	{
		Achievement: models.Achievement{
			Code: "ini-philanthropist",
			Name: "Philanthropist",
			Desc: "You have helped 50 initiatives",
		},
		trigger: func(s State) bool {
			return s.Initiatives >= 50
		},
		completion: func(s State) float64 {
			return math.Min(float64(s.Initiatives)/50, 1)
		},
	},

	// Credits
	{
		Achievement: models.Achievement{
			Code: "crd-gatherer",
			Name: "Gatherer",
			Desc: "You have received your first credits!",
		},
		trigger: func(s State) bool {
			return s.Credits >= 1.0
		},
		completion: func(s State) float64 {
			return math.Min(s.Credits, 1)
		},
	},
	{
		Achievement: models.Achievement{
			Code: "crd-hoarder",
			Name: "Hoarder",
			Desc: "You have received 5000 credits!",
		},
		trigger: func(s State) bool {
			return s.Credits >= 5000.0
		},
		completion: func(s State) float64 {
			return math.Min(s.Credits/5000, 1)
		},
	},
	{
		Achievement: models.Achievement{
			Code: "crd-treasure-master",
			Name: "Treasure Master",
			Desc: "You have received 10.000 credits!",
		},
		trigger: func(s State) bool {
			return s.Credits >= 10_000.0
		},
		completion: func(s State) float64 {
			return math.Min(s.Credits/10_000, 1)
		},
	},
}
