package time2

import (
	"fmt"
	"time"
)

const (
	LayoutDate          = "2006-01-02"
	LayoutDateShort     = "20060102"
	LayoutDateTime      = "2006-01-02 15:04:05"
	LayoutDateTimeShort = "20060102150405"

	TimeDay      = time.Hour * 24
	TimeThreeDay = TimeDay * 3
	TimeWeek     = TimeDay * 7
	TimeMonth    = TimeDay * 30
	TimeYear     = TimeDay * 365

	DaySeconds = 86400
)

func Now(loc ...*time.Location) time.Time {
	_loc := time.Local

	if len(loc) > 0 {
		_loc = loc[0]
	}

	return time.Now().In(_loc)
}

func NowMS(loc ...*time.Location) int64 {
	return Now(loc...).UnixNano() / 1e6
}

func DayStart(t time.Time, loc ...*time.Location) time.Time {
	_loc := time.Local

	if len(loc) > 0 {
		_loc = loc[0]
	}

	locDate := Format(t, LayoutDateShort, _loc)
	locDaystart, _ := time.ParseInLocation(LayoutDateShort, locDate, _loc)
	return locDaystart
}

func Format(t time.Time, layout string, loc ...*time.Location) string {
	_loc := time.Local

	if len(loc) > 0 {
		_loc = loc[0]
	}

	return t.In(_loc).Format(layout)
}

func Date(t time.Time, loc ...*time.Location) string {
	return Format(t, LayoutDate, loc...)
}

func DateShort(t time.Time, loc ...*time.Location) string {
	return Format(t, LayoutDateShort, loc...)
}

func DateTime(t time.Time, loc ...*time.Location) string {
	return Format(t, LayoutDateTime, loc...)
}

func DateTimeShort(t time.Time, loc ...*time.Location) string {
	return Format(t, LayoutDateTimeShort, loc...)
}

func OffsetTS(offset int64) string {
	return fmt.Sprintf("%02d:%02d:%02d", (offset%86400)/3600, (offset%3600)/60, offset%60)
}
