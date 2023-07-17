package types

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// TimeTZFormat is the layout used by the TimeTZ type to convert to and from
// time.Time, and encodes a time of day in hours, minutes and the time zone
// offset (hours and minutes, RFC 3339).
const TimeTZFormat = "15:04Z07:00"

// TimeTz type represents a time of day in the format given by TimeTZFormat.
type TimeTZ string

// Scan implements sql.Scanner so that TimeTZs can be read from a database.
// Expects a data type that maps to string.
func (t *TimeTZ) Scan(src any) error {
	if sSrc, ok := src.(string); ok {
		if len(sSrc) == 11 {
			sSrc += ":00"
		}

		// The database also stores seconds, which we can ignore.
		tTime, err := time.Parse("15:04:05Z07:00", sSrc)
		*t = TimeTZ(tTime.Format(TimeTZFormat))
		return err
	}

	return fmt.Errorf("unsupported type %T", src)
}

// Value implements sql.Valuer so that TimeTZs can be written to a database.
func (t TimeTZ) Value() (driver.Value, error) {
	return string(t), nil
}

// MarshalJSON implements the json.Marshaler interface.
// The timetz is a quoted string in the format given by TimeTZFormat.
func (t TimeTZ) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(TimeTZFormat)+2)
	b = append(b, '"')
	b = append(b, []byte(t)...)
	b = append(b, '"')
	return b, nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// The timetz is expected to be a quoted string in the format given by
// TimeTZFormat.
func (t *TimeTZ) UnmarshalJSON(src []byte) error {
	timeT, err := time.Parse(
		TimeTZFormat,
		string(src[1:len(src)-1]),
	)
	*t = TimeTZ(timeT.Format(TimeTZFormat))
	return err
}

// GormDataType implements gorm's schema.DataTypeInterface, so that gorm can
// establish the database column type of TimeTZs.
func (TimeTZ) GormDataType() string {
	return "time with time zone"
}

// Time returns the TimeTZ parsed into time.Time.
func (t TimeTZ) Time() time.Time {
	tTime, _ := time.Parse(TimeTZFormat, string(t))
	return tTime
}

func (t TimeTZ) String() string {
	return string(t)
}

func TimeTZPtr(v string) *TimeTZ {
	t := TimeTZ(v)
	return &t
}
