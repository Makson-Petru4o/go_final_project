package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const DateFormat = "20060102"

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	start, err := time.Parse(DateFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("invalid start date: %w", err)
	}
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return "", fmt.Errorf("repeat rule is empty")
	}
	switch parts[0] {
	case "d":
		if len(parts) != 2 {
			return "", fmt.Errorf("expected d N")
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil || days < 1 || days > 400 {
			return "", fmt.Errorf("day interval must be from 1 to 400")
		}
		return nextInterval(now, start, 0, days)
	case "y":
		if len(parts) != 1 {
			return "", fmt.Errorf("expected y without parameters")
		}
		return nextInterval(now, start, 1, 0)
	case "w":
		if len(parts) != 2 {
			return "", fmt.Errorf("expected w weekdays")
		}
		days, err := parseList(parts[1], 1, 7)
		if err != nil {
			return "", err
		}
		return nextMatching(now, start, func(date time.Time) bool {
			weekday := (int(date.Weekday())+6)%7 + 1
			return days[weekday]
		})
	case "m":
		if len(parts) < 2 || len(parts) > 3 {
			return "", fmt.Errorf("expected m days [months]")
		}
		days, err := parseList(parts[1], -2, 31)
		if err != nil {
			return "", err
		}
		if days[0] {
			return "", fmt.Errorf("month day cannot be zero")
		}
		var months map[int]bool
		if len(parts) == 3 {
			months, err = parseList(parts[2], 1, 12)
			if err != nil {
				return "", err
			}
		}
		return nextMatching(now, start, func(date time.Time) bool {
			if months != nil && !months[int(date.Month())] {
				return false
			}
			last := time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
			return days[date.Day()] || days[-1] && date.Day() == last || days[-2] && date.Day() == last-1
		})
	default:
		return "", fmt.Errorf("unknown repeat rule: %s", parts[0])
	}
}

func parseList(value string, min, max int) (map[int]bool, error) {
	values := make(map[int]bool)
	for _, item := range strings.Split(value, ",") {
		n, err := strconv.Atoi(item)
		if err != nil || n < min || n > max {
			return nil, fmt.Errorf("invalid repeat value: %q", item)
		}
		values[n] = true
	}
	return values, nil
}

func nextInterval(now, start time.Time, years, days int) (string, error) {
	for date := start.AddDate(years, 0, days); date.Year() <= 9999; date = date.AddDate(years, 0, days) {
		if date.After(now) {
			return date.Format(DateFormat), nil
		}
	}
	return "", fmt.Errorf("next date is outside YYYYMMDD range")
}

func nextMatching(now, start time.Time, matches func(time.Time) bool) (string, error) {
	date := now.AddDate(0, 0, 1)
	if start.After(now) {
		date = start
	}
	limit := date.AddDate(400, 0, 0)
	for ; date.Before(limit) && date.Year() <= 9999; date = date.AddDate(0, 0, 1) {
		if matches(date) {
			return date.Format(DateFormat), nil
		}
	}
	return "", fmt.Errorf("repeat rule has no matching date in YYYYMMDD range")
}
