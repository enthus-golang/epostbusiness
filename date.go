package epostbusiness

import (
	"errors"
	"time"
)

const TimeFormat = "2006-01-02T15:04:05.999"

type Time struct {
	time.Time
}

func (t Time) MarshalJSON() ([]byte, error) {
	if y := t.Time.Year(); y < 0 || y >= 10000 {
		// RFC 3339 is clear that years are 4 digits exactly.
		// See golang.org/issue/4556#c15 for more discussion.
		return nil, errors.New("Time.MarshalJSON: year outside of range [0,9999]")
	}

	b := make([]byte, 0, len(TimeFormat)+2)
	b = append(b, '"')
	b = t.Time.AppendFormat(b, TimeFormat)
	b = append(b, '"')

	return b, nil
}

func (t *Time) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" {
		return nil
	}
	// Fractional seconds are handled implicitly by Parse.
	var err error
	t.Time, err = time.ParseInLocation(`"`+TimeFormat+`"`, string(data), time.Local)
	return err
}
