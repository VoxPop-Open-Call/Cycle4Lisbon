package seeders

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/types"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/random"
	"gorm.io/gorm"
)

type initiative struct{}

var Initiative initiative

func (initiative) Seed(db *gorm.DB) {
	var institutions []models.Institution
	db.Find(&institutions)

	var sdgs []models.SDG
	db.Find(&sdgs)

	var initiatives = []models.Initiative{
		{
			Title:       "Initiative 0",
			Description: "Quia nihil deleniti esse minus sit hic.",
			Goal:        7000,
			EndDate:     types.Date("2050-10-10"),
			Enabled:     true,
			Institution: institutions[0],
			Sponsors:    random.PickFrom(institutions),
			SDGs:        random.PickFrom(sdgs),
		},
		{
			Title:       "Initiative 1",
			Description: "Est distinctio odit quis ratione illum.",
			Goal:        5000,
			EndDate:     types.Date("2036-01-10"),
			Enabled:     true,
			Institution: institutions[1],
			Sponsors:    random.PickFrom(institutions),
			SDGs:        random.PickFrom(sdgs),
		},
		{
			Title:       "Initiative 3",
			Description: "Ex cumque iure sed aut assumenda.",
			Goal:        5000,
			EndDate:     types.Date("2036-01-10"),
			Enabled:     false,
			Institution: institutions[2],
			Sponsors:    random.PickFrom(institutions),
			SDGs:        random.PickFrom(sdgs),
		},
	}

	db.CreateInBatches(initiatives, 50)
}
