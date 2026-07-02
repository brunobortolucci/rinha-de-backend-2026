package vector

import (
	"errors"
	"time"
)

var errInvalidTimestamp = errors.New("invalid timestamp format")

func parseUTC(s string) (time.Time, bool) {
	if len(s) != 20 || s[19] != 'Z' || s[4] != '-' || s[7] != '-' || s[10] != 'T' || s[13] != ':' || s[16] != ':' {
		return time.Time{}, false
	}

	year, ok := atoi4(s, 0)
	if !ok {
		return time.Time{}, false
	}
	month, ok := atoi2(s, 5)
	if !ok || month < 1 || month > 12 {
		return time.Time{}, false
	}
	day, ok := atoi2(s, 8)
	if !ok || day < 1 || day > 31 {
		return time.Time{}, false
	}
	hour, ok := atoi2(s, 11)
	if !ok || hour > 23 {
		return time.Time{}, false
	}
	minute, ok := atoi2(s, 14)
	if !ok || minute > 59 {
		return time.Time{}, false
	}
	second, ok := atoi2(s, 17)
	if !ok || second > 59 {
		return time.Time{}, false
	}

	return time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC), true
}

func atoi2(s string, off int) (int, bool) {
	d0 := s[off] - '0'
	d1 := s[off+1] - '0'
	if d0 > 9 || d1 > 9 {
		return 0, false
	}
	return int(d0)*10 + int(d1), true
}

func atoi4(s string, off int) (int, bool) {
	d0 := s[off] - '0'
	d1 := s[off+1] - '0'
	d2 := s[off+2] - '0'
	d3 := s[off+3] - '0'
	if d0 > 9 || d1 > 9 || d2 > 9 || d3 > 9 {
		return 0, false
	}
	return int(d0)*1000 + int(d1)*100 + int(d2)*10 + int(d3), true
}
