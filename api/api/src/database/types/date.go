package types

import (
	"database/sql"
	"database/sql/driver"
	"time"
)

// DateFormat is the layout used by the Date type to convert to and from
// time.Time.
const DateFormat = time.DateOnly

// Date type represents a date in the format given by DateFormat.
type Date string

// Scan implements sql.Scanner so that Dates can be read from a database.
// Expects a data type that maps to time.Time.
func (date *Date) Scan(src any) error {
	nullTime := &sql.NullTime{}
	err := nullTime.Scan(src)
	*date = Date(nullTime.Time.Format(DateFormat))
	return err
}

// Value implements sql.Valuer so that Dates can be written to a database.
func (date Date) Value() (driver.Value, error) {
	return time.Parse(DateFormat, string(date))
}

// MarshalJSON implements the json.Marshaler interface.
// The date is a quoted string in the format given by DateFormat.
func (date Date) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(DateFormat)+2)
	b = append(b, '"')
	b = append(b, []byte(date)...)
	b = append(b, '"')
	return b, nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// The date is expected to be a quoted string in the format given by
// DateFormat.
func (date *Date) UnmarshalJSON(src []byte) error {
	t, err := time.Parse(
		DateFormat,
		string(src[1:len(src)-1]),
	)
	*date = Date(t.Format(DateFormat))
	return err
}

// GormDataType implements gorm's schema.DataTypeInterface, so that gorm can
// establish the database column type of Dates.
func (Date) GormDataType() string {
	return "date"
}

// Time returns the Date parsed into time.Time.
func (date Date) Time() time.Time {
	tDate, _ := time.Parse(DateFormat, string(date))
	return tDate
}

func DatePtr(v string) *Date {
	d := Date(v)
	return &d
}
