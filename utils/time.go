//go:build !test
// +build !test

package utils

import (
	"database/sql/driver"
	"fmt"
	"time"
)

const (
	RFC3339WithoutTimezone = "2006-01-02T15:04:05"
)

func genNiceTime() string {
	timeFormat := "Mon 2 Jan 2006 15-04-05"
	time := time.Now()
	return time.Format(timeFormat)
}

func Now() string {
	return time.Now().Format(time.RFC3339)
}

func NewTime() Time {
	return Time{
		time.Now(),
	}
}

type Time struct {
	time.Time
}

func (t Time) Value() (driver.Value, error) {
	return t.Format(RFC3339WithoutTimezone), nil
}

func (t *Time) Scan(value interface{}) error {
	if value == nil {
		*t = Time{Time: time.Time{}}
		return nil
	}

	// Convert the value to time.Time
	switch v := value.(type) {
	case time.Time:
		*t = Time{Time: v}
		return nil
	case []byte:
		parsedTime, err := time.Parse(time.RFC3339, string(v))
		if err != nil {
			return err
		}
		*t = Time{Time: parsedTime}
		return nil
	case string:
		parsedTime, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return err
		}
		*t = Time{Time: parsedTime}
		return nil
	default:
		return fmt.Errorf("unsupported Scan type for utils.Time: %T", value)
	}
}
