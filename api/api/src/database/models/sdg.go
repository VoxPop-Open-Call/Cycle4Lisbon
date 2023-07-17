package models

import (
	"fmt"
	"log"

	"bitbucket.org/pensarmais/cycleforlisbon/resources"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SDG struct {
	Code        int    `json:"code" gorm:"primaryKey" binding:"required"`
	Title       string `json:"title" gorm:"unique;not null"`
	Description string `json:"description"`
	ImageURI    string `json:"imageURI" gorm:"not null"`
}

func (SDG) Migrate(db *gorm.DB) error {
	for _, sdg := range resources.SDGs {
		if err := db.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&SDG{
			Code:        sdg.Code,
			Title:       sdg.Title,
			Description: sdg.Description,
			ImageURI:    fmt.Sprintf("/public/assets/sdg/E-WEB-Goal-%02d.png", sdg.Code),
		}).Error; err != nil {
			log.Printf("error creating SDG %+v: %v", sdg, err)
		}
	}

	return nil
}
