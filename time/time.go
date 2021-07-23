package time

import (
    "errors"
    "fmt"
    "time"
)

type Time time.Time

const DefaultFormat = time.RFC3339

var layouts = []string{
    DefaultFormat,
    "2006-01-02T15:04Z",        // ISO 8601 UTC
    "2006-01-02T15:04:05Z",     // ISO 8601 UTC
    "2006-01-02T15:04:05.000Z", // ISO 8601 UTC
    "2006-01-02T15:04:05",      // ISO 8601 UTC
    "2006-01-02 15:04",         // Custom UTC
    "2006-01-02 15:04:05",      // Custom UTC
    "2006-01-02 15:04:05.000",  // Custom UTC
}

func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`%s`,t.String())), nil
}

func (t *Time) UnmarshalJSON(b []byte) (error) {
    timeString := string(b)
    for _, layout := range layouts {
        tt, err := time.Parse(layout, timeString)
        if err == nil {
            *t = Time(tt)
            return nil
        }
    }
    return errors.New(fmt.Sprintf("Invalid date format: %s", timeString))
}

func (t Time) Unix() int64 {
    return time.Time(t).Unix()
}

func (t Time) Time() time.Time {
    return time.Time(t).UTC()
}

func (t *Time) String() string {
    tt := time.Time(*t)
    return tt.Format(DefaultFormat)
}

