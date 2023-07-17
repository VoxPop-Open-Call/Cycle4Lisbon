package seeders

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/random"
	"gorm.io/gorm"
)

type institution struct{}

var Institution institution

func (institution) Seed(db *gorm.DB) {

	var institutions = []models.Institution{
		{
			Name:        "shopify",
			Description: random.AlphanumericString(30),
		},
		{
			Name:        "loom",
			Description: random.AlphanumericString(30),
		},
		{
			Name:        "Unsplash",
			Description: random.AlphanumericString(30),
		},
		{
			Name:        "CM Lisboa",
			Description: random.AlphanumericString(30),
		},
		{
			Name:        "IKEA",
			Description: random.AlphanumericString(50),
		},
		{
			Name:        "Santa Casa",
			Description: random.AlphanumericString(50),
		},
		{
			Name:        "Lisbon Project",
			Description: random.AlphanumericString(50),
		},
		{
			Name:        "Banco Alimentar",
			Description: random.AlphanumericString(50),
		},
	}

	db.CreateInBatches(institutions, 50)
}
