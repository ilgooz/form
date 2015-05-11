package form

import "time"

func Time(ts string) (time.Time, error) {
	t := time.Time{}
	err := t.UnmarshalText([]byte(ts))
	return t, err
}
