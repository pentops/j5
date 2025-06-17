package date_j5t

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func NewDate(year, month, day int32) *Date {
	return &Date{
		Year:  year,
		Month: month,
		Day:   day,
	}
}

// AsTime returns a time.Time object with the date set to the Date value at
// midnight in the morning in the specified location.
func (dd *Date) AsTime(location *time.Location) time.Time {
	return time.Date(
		int(dd.Year),
		time.Month(dd.Month),
		int(dd.Day), 0, 0, 0, 0, location)
}

// DateString returns an ISO 8601 / RFC3339 date string.
func (dd *Date) DateString() string {
	return fmt.Sprintf("%4d-%02d-%02d", dd.Year, dd.Month, dd.Day)
}

func (dd *Date) MarshalText() ([]byte, error) {
	str := dd.DateString()
	return []byte(str), nil
}

func (dd *Date) UnmarshalText(data []byte) error {
	got, err := DateFromString(string(data))
	if err != nil {
		return err
	}
	dd.Year = got.Year
	dd.Month = got.Month
	dd.Day = got.Day
	return nil
}

func (dd *Date) Value() (driver.Value, error) {
	if dd == nil {
		return nil, nil
	}
	return dd.DateString(), nil
}

func (dd *Date) Scan(src any) error {
	if src == nil {
		dd.Year, dd.Month, dd.Day = 0, 0, 0
		return nil
	}
	if str, ok := src.(string); ok {
		got, err := DateFromString(str)
		if err != nil {
			return fmt.Errorf("failed to parse date from string: %w", err)
		}
		dd.Year = got.Year
		dd.Month = got.Month
		dd.Day = got.Day
		return nil
	}
	return fmt.Errorf("cannot scan type %T into Date", src)
}

func DateFromString(data string) (*Date, error) {
	parts := strings.Split(data, "-")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid date string: %s", data)
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid date string: %s", data)
	}

	month, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid date string: %s", data)
	}

	day, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid date string: %s", data)
	}

	dd := &Date{
		Year:  int32(year),
		Month: int32(month),
		Day:   int32(day),
	}

	return dd, nil
}

// Equals returns true if the two dates are equal.
func (dd *Date) Equals(other *Date) bool {
	return dd.Year == other.Year && dd.Month == other.Month && dd.Day == other.Day
}

// DateFromTime returns a Date object from a time.Time object, in the timezone of the
// time.Time
func DateFromTime(tt time.Time) *Date {
	y, m, d := tt.Date()
	return &Date{
		Year:  int32(y),
		Month: int32(m),
		Day:   int32(d),
	}
}

// DateFromTimeIn returns a Date object from a time.Time object, in the specified
// timezone.
func DateFromTimeIn(tt time.Time, loc *time.Location) *Date {
	return DateFromTime(tt.In(loc))
}

func (dd *Date) AddDate(years, months, days int32) *Date {
	// The go implementation of this is fairly efficient, even if we are wasting
	// the time parts. Still, this could be improved.
	tt := dd.AsTime(time.UTC).AddDate(int(years), int(months), int(days))
	return DateFromTime(tt)
}

func (dd *Date) Before(other *Date) bool {
	return dd.AsTime(time.UTC).Before(other.AsTime(time.UTC))
}

func (dd *Date) After(other *Date) bool {
	return dd.AsTime(time.UTC).After(other.AsTime(time.UTC))
}

func (dd *Date) IsZero() bool {
	return dd.Year == 0 && dd.Month == 0 && dd.Day == 0
}
