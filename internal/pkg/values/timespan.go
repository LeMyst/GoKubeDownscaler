package values

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	errInvalidWeekday          = errors.New("error: specified weekday is invalid")
	errRelativeTimespanInvalid = errors.New("error: specified relative timespan is invalid")
)

type TimeSpan interface {
	// inTimeSpan checks if time is in the timespan or not
	isTimeInSpan(time.Time) bool
}

type timeSpans []TimeSpan

// inTimeSpans checks if current time is in one of the timespans or not
func (t *timeSpans) inTimeSpans() bool {
	for _, timespan := range *t {
		if !timespan.isTimeInSpan(time.Now()) {
			continue
		}
		return true
	}
	return false
}

func (t *timeSpans) Set(value string) error {
	spans := strings.Split(value, ",")
	var timespans []TimeSpan
	for _, timespanText := range spans {
		timespanText = strings.TrimSpace(timespanText)

		if isAbsoluteTimestamp(timespanText) {
			// parse as absolute timestamp
			timespan, err := parseAbsoluteTimeSpan(timespanText)
			if err != nil {
				return fmt.Errorf("failed to parse absolute timespan: %w", err)
			}
			timespans = append(timespans, timespan)
			continue
		}

		// parse as relative timestamp
		timespan, err := parseRelativeTimeSpan(timespanText)
		if err != nil {
			return fmt.Errorf("failed to parse relative timespan: %w", err)
		}
		timespans = append(timespans, timespan)
	}
	*t = timeSpans(timespans)
	return nil
}

func (t *timeSpans) String() string {
	return fmt.Sprint(*t)
}

func parseAbsoluteTimeSpan(timespan string) (absoluteTimeSpan, error) {
	timestamps := strings.Split(timespan, " - ")
	fromTime, err := time.Parse(time.RFC3339, timestamps[0])
	if err != nil {
		return absoluteTimeSpan{}, fmt.Errorf("failed to parse rfc3339 timestamp: %w", err)
	}
	toTime, err := time.Parse(time.RFC3339, timestamps[1])
	if err != nil {
		return absoluteTimeSpan{}, fmt.Errorf("failed to parse rfc3339 timestamp: %w", err)
	}

	return absoluteTimeSpan{
		from: fromTime,
		to:   toTime,
	}, nil
}

func parseRelativeTimeSpan(timespanString string) (*relativeTimeSpan, error) {
	timespan := relativeTimeSpan{}

	parts := strings.Split(timespanString, " ")
	if len(parts) != 3 {
		return nil, errRelativeTimespanInvalid
	}

	weekdaySpan := strings.Split(parts[0], "-")
	if len(weekdaySpan) != 2 {
		return nil, errRelativeTimespanInvalid
	}
	timeSpan := strings.Split(parts[1], "-")
	if len(timeSpan) != 2 {
		return nil, errRelativeTimespanInvalid
	}
	timezone := parts[2]

	var err error
	timespan.timezone, err = time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timezone: %w", err)
	}
	timespan.timeFrom, err = time.ParseInLocation("15:04", timeSpan[0], timespan.timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to parse 'timeFrom': %w", err)
	}
	timespan.timeTo, err = time.ParseInLocation("15:04", timeSpan[1], timespan.timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to parse 'timeTo': %w", err)
	}
	timespan.weekdayFrom, err = getWeekday(weekdaySpan[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse 'weekdayFrom': %w", err)
	}
	timespan.weekdayTo, err = getWeekday(weekdaySpan[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse 'weekdayTo': %w", err)
	}

	return &timespan, nil
}

type relativeTimeSpan struct {
	timezone    *time.Location
	weekdayFrom time.Weekday
	weekdayTo   time.Weekday
	timeFrom    time.Time
	timeTo      time.Time
}

// isWeekdayInRange checks if the weekday falls into the weekday range
func (t relativeTimeSpan) isWeekdayInRange(weekday time.Weekday) bool {
	if t.weekdayFrom <= t.weekdayTo { // check if range wraps across weeks
		return weekday >= t.weekdayFrom && weekday <= t.weekdayTo
	}
	return weekday >= t.weekdayFrom || weekday <= t.weekdayTo
}

// isTimeOfDayInRange checks if the time falls into the time of day range
func (t relativeTimeSpan) isTimeOfDayInRange(timeOfDay time.Time) bool {
	if t.timeFrom.After(t.timeTo) { // check if range wraps across days
		return timeOfDay.After(t.timeFrom) || timeOfDay.Equal(t.timeFrom) || timeOfDay.Before(t.timeTo)
	}
	return (t.timeFrom.Before(timeOfDay) || t.timeFrom.Equal(timeOfDay)) && t.timeTo.After(timeOfDay)
}

// isTimeInSpan checks if the time is in the span
func (t relativeTimeSpan) isTimeInSpan(targetTime time.Time) bool {
	targetTime = targetTime.In(t.timezone)
	timeOfDay := getTimeOfDay(targetTime)
	weekday := targetTime.Weekday()
	return t.isTimeOfDayInRange(timeOfDay) && t.isWeekdayInRange(weekday)
}

// inLocation returns an array of relative timespans matching the timespan converted to the given location
func (t relativeTimeSpan) inLocation(timezone *time.Location) []relativeTimeSpan {
	var result []relativeTimeSpan
	sameDays := relativeTimeSpan{
		timezone:    timezone,
		weekdayFrom: t.weekdayFrom,
		weekdayTo:   t.weekdayTo,
		timeFrom:    t.timeFrom.In(timezone),
		timeTo:      t.timeTo.In(timezone),
	}
	result = append(result, sameDays)
	if isTimeFromSkippedToPreviousDay(sameDays.timeFrom) {
		daysBefore := relativeTimeSpan{
			timezone:    timezone,
			timeFrom:    sameDays.timeFrom.Add(24 * time.Hour),
			timeTo:      sameDays.timeTo.Add(24 * time.Hour),
			weekdayFrom: sameDays.weekdayFrom - 1,
			weekdayTo:   sameDays.weekdayTo - 1,
		}
		result = append(result, daysBefore)
	}
	if isTimeToSkippedToNextDay(sameDays.timeTo) {
		daysAfter := relativeTimeSpan{
			timezone:    timezone,
			timeFrom:    sameDays.timeFrom.Add(-24 * time.Hour),
			timeTo:      sameDays.timeTo.Add(-24 * time.Hour),
			weekdayFrom: sameDays.weekdayFrom + 1,
			weekdayTo:   sameDays.weekdayTo + 1,
		}
		result = append(result, daysAfter)
	}
	return result
}

// Check if timeFrom skipped to the previous day
func isTimeFromSkippedToPreviousDay(timeFrom time.Time) bool {
	return timeFrom.Year() == -1
}

// Check if timeTo skipped to the following day
func isTimeToSkippedToNextDay(timeTo time.Time) bool {
	return asExclusiveTimestamp(timeTo).Day() == 2
}

type absoluteTimeSpan struct {
	from time.Time
	to   time.Time
}

// isTimeInSpan check if the time is in the span
func (t absoluteTimeSpan) isTimeInSpan(targetTime time.Time) bool {
	return (t.from.Before(targetTime) || t.from.Equal(targetTime)) && t.to.After(targetTime)
}

// isAbsoluteTimestamp checks if timestamp string is absolute
func isAbsoluteTimestamp(timestamp string) bool {
	return strings.Contains(timestamp, " - ")
}

// getWeekday gets the weekday from the given string
func getWeekday(weekday string) (time.Weekday, error) {
	weekdays := map[string]time.Weekday{
		"sun": time.Sunday,
		"mon": time.Monday,
		"tue": time.Tuesday,
		"wed": time.Wednesday,
		"thu": time.Thursday,
		"fri": time.Friday,
		"sat": time.Saturday,
	}

	if day, ok := weekdays[strings.ToLower(weekday)]; ok {
		return day, nil
	}

	return 0, errInvalidWeekday
}

// getTimeOfDay gets the time of day of the given time
func getTimeOfDay(targetTime time.Time) time.Time {
	return time.Date(0, time.January, 1,
		targetTime.Hour(),
		targetTime.Minute(),
		targetTime.Second(),
		targetTime.Nanosecond(),
		targetTime.Location(),
	)
}

// doTimespansOverlap checks if the given timespans overlap with each other
func doTimespansOverlap(span1, span2 TimeSpan) bool {
	switch s1 := span1.(type) {
	case absoluteTimeSpan:
		if s2, ok := span2.(absoluteTimeSpan); ok {
			return absAndAbsOverlap(s1, s2)
		}
		return relAndAbsOverlap(span2.(relativeTimeSpan), s1)
	case relativeTimeSpan:
		if s2, ok := span2.(absoluteTimeSpan); ok {
			return relAndAbsOverlap(s1, s2)
		}
		return relAndRelOverlap(s1, span2.(relativeTimeSpan))
	}
	return false // this shouldn't ever be reached
}

// relAndRelOverlap checks if two relative timespans overlap
func relAndRelOverlap(rel1 relativeTimeSpan, rel2 relativeTimeSpan) bool {
	_, tzOffset1 := rel1.timeFrom.Zone()
	_, tzOffset2 := rel2.timeFrom.Zone()
	if tzOffset1 == tzOffset2 { // optimized path for timespans with same timezone offset
		return relAndRelSameTimezoneOverlap(rel1, rel2)
	}

	// slow path for timespans with different timezones
	rel2List := rel2.inLocation(rel1.timezone)
	for _, rel2 := range rel2List {
		if relAndRelSameTimezoneOverlap(rel1, rel2) {
			return true
		}
	}
	return false
}

// relAndRelSameTimezoneOverlap checks if two relative timespans overlap, without converting them to the same timezone before
func relAndRelSameTimezoneOverlap(rel1 relativeTimeSpan, rel2 relativeTimeSpan) bool {
	if rel1.timeFrom.After(rel1.timeTo) && rel2.timeFrom.After(rel2.timeTo) { // if both timespans are reversed, they both overlap anytime
		return true
	}
	overlappingWeekdays := rel1.isWeekdayInRange(rel2.weekdayFrom) ||
		rel1.isWeekdayInRange(rel2.weekdayTo) ||
		rel2.isWeekdayInRange(rel1.weekdayFrom) ||
		rel2.isWeekdayInRange(rel1.weekdayTo)
	if !overlappingWeekdays {
		return false
	}
	overlappingTimeOfDays := rel1.isTimeOfDayInRange(rel2.timeFrom) ||
		rel1.isTimeOfDayInRange(asExclusiveTimestamp(rel2.timeTo)) ||
		rel2.isTimeOfDayInRange(rel1.timeFrom) ||
		rel2.isTimeOfDayInRange(asExclusiveTimestamp(rel1.timeTo))
	return overlappingTimeOfDays
}

// relAndAbsOverlap checks if a relative and an aboslute timespan overlap
func relAndAbsOverlap(rel relativeTimeSpan, abs absoluteTimeSpan) bool {
	abs.from = abs.from.In(rel.timezone)
	abs.to = abs.to.In(rel.timezone)
	if abs.to.Sub(abs.from) >= 7*24*time.Hour {
		return true
	}
	if rel.isTimeInSpan(abs.from) || rel.isTimeInSpan(asExclusiveTimestamp(abs.to)) {
		return true
	}
	if rel.weekdayFrom > rel.weekdayTo { // check if weekdays are in reverse
		return abs.from.Weekday() > rel.weekdayFrom || abs.from.Weekday() < rel.weekdayTo ||
			abs.to.Weekday() > rel.weekdayFrom || abs.to.Weekday() < rel.weekdayTo
	}
	return rel.weekdayFrom > abs.from.Weekday() && rel.weekdayFrom < abs.to.Weekday() ||
		rel.weekdayTo > abs.from.Weekday() && rel.weekdayTo < abs.to.Weekday()
}

// absAndAbsOverlap checks if two absolute timespans overlap
func absAndAbsOverlap(abs1 absoluteTimeSpan, abs2 absoluteTimeSpan) bool {
	return abs1.from.Before(abs2.to) && abs1.to.After(abs2.from)
}
