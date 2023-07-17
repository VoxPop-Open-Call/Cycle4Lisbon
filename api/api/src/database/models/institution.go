package models

type Institution struct {
	BaseModel
	Name        string `json:"name" gorm:"unique;not null"`
	Description string `json:"description"`
}
