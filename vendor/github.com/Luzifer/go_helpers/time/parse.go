package time

import (
	"errors"
	gotime "time"
)

// ErrParseNotPossible describes a collected error which is returned
// when the given date value could not be parsed with any given layout
var ErrParseNotPossible = errors.New("No layout matched given value or other error occurred")

// MultiParse takes multiple layout strings and tries to parse the
// value with them. In case none of the layouts matches or another error
// occurs ErrParseNotPossible is returned
func MultiParse(layouts []string, value string) (gotime.Time, error) {
	for _, layout := range layouts {
		if t, err := gotime.Parse(layout, value); err == nil {
			return t, nil
		}
	}

	return gotime.Time{}, ErrParseNotPossible
}

// MultiParseInLocation takes multiple layouts and tries to parse the
// value with them using time.ParseInLocation. In case none of the layouts
// matches or another error occurs ErrParseNotPossible is returned
func MultiParseInLocation(layouts []string, value string, loc *gotime.Location) (gotime.Time, error) {
	for _, layout := range layouts {
		if t, err := gotime.ParseInLocation(layout, value, loc); err == nil {
			return t, nil
		}
	}

	return gotime.Time{}, ErrParseNotPossible
}
