package models

import (
	"database/sql/driver"
	"fmt"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/types"
)

// ExternalContent represents events, news articles, etc., from external
// sources. Diferent content types will omit some fields.
//
// See package `scraper`, where the content is retrieved and parsed.
type ExternalContent struct {
	BaseModel
	// Type of the entry, for example "news" or "event".
	Type string `json:"type" example:"news"`

	State ExternalContentState `json:"state" gorm:"type:varchar(8);not null;default:pending" binding:"oneof=pending approved rejected" example:"approved"`

	Title      string `json:"title"`
	Subtitle   string `json:"subtitle"`
	ImageUrl   string `json:"imageUrl"`
	ArticleUrl string `json:"articleUrl" gorm:"unique;not null"`
	// Subject is a list of tags separated by a comma.
	// For example, "visitas guiadas", "artes" or "direitos sociais, educação".
	Subject string `json:"subject,omitempty"`

	// Description of the event, truncated to around 200 bytes (less than 200
	// characters).
	//
	// The string may contain html tags which, because of the blind truncation,
	// may not be correctly closed or even complete. This value is probably
	// useless.
	//
	// Only content of type `event` contains this field.
	Description string `json:"description,omitempty"`

	// Period includes the start and end dates of the event, in the format "12
	// dezembro 2022 a 31 dezembro 2023".
	//
	// All the values I encountered are in this format, but it's probably not a
	// good idea to count on it to parse them.
	//
	// Only content of type `event` contains this field.
	Period string `json:"period,omitempty"`

	// Time is a free-form description of the time of day at which the event
	// occurs, for example: "sex: 21h30; sáb: 19h" or "vários horários".
	//
	// Only content of type `event` contains this field.
	Time string `json:"time,omitempty"`

	// Date of the news article.
	//
	// Only content of type `news` contains this field.
	Date *types.Date `json:"date,omitempty" gorm:"default:null"`

	LanguageCode string   `json:"languageCode" gorm:"not null"`
	Language     Language `json:"language" gorm:"foreignKey:LanguageCode;references:Code"`
}

// ExternalContentState can be empty, "pending", "approved" or "rejected".
type ExternalContentState string

// Scan implements sql.Scanner so that ExternalContentStates can be read from a
// database.
// Database types that map to string and []byte are supported.
func (s *ExternalContentState) Scan(src any) error {
	var val string
	if s, ok := src.(string); ok {
		val = s
	} else if b, ok := src.([]byte); ok {
		val = string(b)
	} else {
		return fmt.Errorf("unable to scan type %T into ExternalContentState", src)
	}

	if !isValidExternalContentState(val) {
		return fmt.Errorf("invalid value for ExternalContentState: %s", val)
	}

	*s = ExternalContentState(val)
	return nil
}

// Value implements sql.Valuer so that ExternalContentStates can be written to
// a database.
// ExternalContentState maps to string.
func (m ExternalContentState) Value() (driver.Value, error) {
	if !isValidExternalContentState(string(m)) {
		return "", fmt.Errorf("invalid value for ExternalContentState: %s", m)
	}
	return string(m), nil
}

func isValidExternalContentState(val string) bool {
	return val == "" ||
		val == "pending" || val == "approved" || val == "rejected"
}
