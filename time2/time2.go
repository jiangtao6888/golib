package time2

import (
	"fmt"
	"time"
)

const (
	DateLayout     = "20060102"
	DtLayout       = "2006-01-02"
	DateTimeLayout = "2006-01-02 15:04:05"

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

func TodayStart(t time.Time, loc ...*time.Location) time.Time {
	_loc := time.Local

	if len(loc) > 0 {
		_loc = loc[0]
	}

	locDate := Format(t, DateLayout, _loc)
	locDaystart, _ := time.ParseInLocation(DateLayout, locDate, _loc)
	return locDaystart
}

func Format(t time.Time, layout string, loc ...*time.Location) string {
	_loc := time.Local

	if len(loc) > 0 {
		_loc = loc[0]
	}

	return t.In(_loc).Format(layout)
}

func FormatD(t time.Time, loc ...*time.Location) string {
	return Format(t, DateLayout, loc...)
}

func FormatDt(t time.Time, loc ...*time.Location) string {
	return Format(t, DtLayout, loc...)
}

func FormatDateTime(t time.Time, loc ...*time.Location) string {
	return Format(t, DateTimeLayout, loc...)
}

func OffsetTS(offset int64) string {
	return fmt.Sprintf("%02d:%02d:%02d", (offset%86400)/3600, (offset%3600)/60, offset%60)
}
