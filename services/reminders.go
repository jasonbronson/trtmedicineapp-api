package services

import (
	"time"

	"github.com/jasonbronson/go-gin-boilerplate/models"
)

type DueReminder struct {
	Medicine models.Medicine `json:"medicine"`
	Schedule models.Schedule `json:"schedule"`
	DueAt    time.Time       `json:"due_at"`
}

func IsScheduleDue(schedule models.Schedule, now time.Time) bool {
	if now.Before(schedule.StartDate) {
		return false
	}
	if schedule.EndDate != nil && now.After(endOfDay(*schedule.EndDate)) {
		return false
	}

	switch schedule.ScheduleType {
	case "daily":
		return timeMatches(schedule, now)
	case "every_other_day":
		return daysBetween(schedule.StartDate, now)%2 == 0 && timeMatches(schedule, now)
	case "every_x_days":
		if schedule.IntervalDays == nil || *schedule.IntervalDays < 1 {
			return false
		}
		return daysBetween(schedule.StartDate, now)%*schedule.IntervalDays == 0 && timeMatches(schedule, now)
	case "weekly":
		interval := 1
		if schedule.WeeklyInterval != nil {
			interval = *schedule.WeeklyInterval
		}
		return daysBetween(schedule.StartDate, now)%7 == 0 && daysBetween(schedule.StartDate, now)/7%interval == 0 && timeMatches(schedule, now)
	case "days_of_week":
		for _, day := range schedule.DaysOfWeek {
			if int(now.Weekday()) == day.Day {
				return timeMatches(schedule, now)
			}
		}
	case "monthly":
		return schedule.MonthlyDay != nil && now.Day() == *schedule.MonthlyDay && timeMatches(schedule, now)
	case "days_of_month":
		for _, day := range schedule.DaysOfMonth {
			if now.Day() == day.Day {
				return timeMatches(schedule, now)
			}
		}
	case "cycle":
		activeDays := 0
		inactiveDays := 0
		if schedule.CycleActiveDays != nil {
			activeDays = *schedule.CycleActiveDays
		}
		if schedule.CycleInactiveDays != nil {
			inactiveDays = *schedule.CycleInactiveDays
		}
		if activeDays == 0 || inactiveDays == 0 {
			return false
		}
		cycleStart := schedule.StartDate
		if schedule.CycleStartDate != nil {
			cycleStart = *schedule.CycleStartDate
		}
		return daysBetween(cycleStart, now)%(activeDays+inactiveDays) < activeDays && timeMatches(schedule, now)
	case "every_x_hours":
		if schedule.IntervalHours == nil || *schedule.IntervalHours < 1 {
			return false
		}
		startAt, ok := scheduleStartAt(schedule, now.Location())
		if !ok || now.Before(startAt) {
			return false
		}
		elapsedMinutes := int(now.Sub(startAt).Minutes())
		return elapsedMinutes >= 0 && elapsedMinutes%(*schedule.IntervalHours*60) == 0
	}
	return false
}

func DueAt(schedule models.Schedule, now time.Time) time.Time {
	if schedule.ScheduleType == "every_x_hours" {
		return now.Truncate(time.Minute)
	}
	if schedule.TimeOfDay == nil {
		return now.Truncate(time.Hour)
	}
	parsed, err := time.Parse("15:04", *schedule.TimeOfDay)
	if err != nil {
		return now
	}
	return time.Date(now.Year(), now.Month(), now.Day(), parsed.Hour(), parsed.Minute(), 0, 0, now.Location())
}

func timeMatches(schedule models.Schedule, now time.Time) bool {
	if schedule.TimeOfDay == nil {
		return false
	}
	return *schedule.TimeOfDay == now.Format("15:04")
}

func scheduleStartAt(schedule models.Schedule, location *time.Location) (time.Time, bool) {
	if schedule.TimeOfDay == nil {
		return time.Time{}, false
	}
	parsed, err := time.Parse("15:04", *schedule.TimeOfDay)
	if err != nil {
		return time.Time{}, false
	}
	return time.Date(schedule.StartDate.Year(), schedule.StartDate.Month(), schedule.StartDate.Day(), parsed.Hour(), parsed.Minute(), 0, 0, location), true
}

func daysBetween(start, end time.Time) int {
	startDate := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	endDate := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())
	return int(endDate.Sub(startDate).Hours() / 24)
}

func endOfDay(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, date.Location())
}
