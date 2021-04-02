package utils

import "time"

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
