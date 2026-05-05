package handlers

import (
	"errors"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jasonbronson/go-gin-boilerplate/config"
	apilog "github.com/jasonbronson/go-gin-boilerplate/library/log"
	"github.com/jasonbronson/go-gin-boilerplate/middleware"
	"github.com/jasonbronson/go-gin-boilerplate/models"
	"github.com/jasonbronson/go-gin-boilerplate/services"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var timeOfDayPattern = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)

type scheduleRequest struct {
	ScheduleType      string  `json:"schedule_type" binding:"required"`
	StartDate         string  `json:"start_date" binding:"required"`
	EndDate           *string `json:"end_date"`
	TimeOfDay         *string `json:"time_of_day"`
	IntervalHours     *int    `json:"interval_hours"`
	IntervalDays      *int    `json:"interval_days"`
	WeeklyInterval    *int    `json:"weekly_interval"`
	MonthlyDay        *int    `json:"monthly_day"`
	CycleActiveDays   *int    `json:"cycle_active_days"`
	CycleInactiveDays *int    `json:"cycle_inactive_days"`
	CycleStartDate    *string `json:"cycle_start_date"`
	DaysOfWeek        []int   `json:"days_of_week"`
	DaysOfMonth       []int   `json:"days_of_month"`
}

func ListSchedules(g *gin.Context) {
	if !medicineExists(g, g.Param("id")) {
		return
	}
	var schedules []models.Schedule
	config.Cfg.GormDB.Preload("DaysOfWeek").Preload("DaysOfMonth").
		Where("medicine_id = ? AND user_id = ?", g.Param("id"), middleware.UserID(g)).
		Find(&schedules)
	g.JSON(http.StatusOK, gin.H{"schedules": schedules})
}

func CreateSchedule(g *gin.Context) {
	medicineID := g.Param("id")
	userID := middleware.UserID(g)
	if !medicineExists(g, medicineID) {
		apilog.ZapWarn("create schedule medicine not found",
			zap.String("user_id", userID),
			zap.String("medicine_id", medicineID),
		)
		return
	}
	var req scheduleRequest
	if err := g.ShouldBindJSON(&req); err != nil {
		apilog.ZapWarn("create schedule bind failed",
			zap.String("user_id", userID),
			zap.String("medicine_id", medicineID),
			zap.Error(err),
		)
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	logScheduleRequest("create schedule request", userID, medicineID, "", req)
	schedule, err := buildSchedule(req, userID, medicineID)
	if err != nil {
		apilog.ZapWarn("create schedule validation failed",
			zap.String("user_id", userID),
			zap.String("medicine_id", medicineID),
			zap.String("schedule_type", req.ScheduleType),
			zap.Ints("days_of_week", req.DaysOfWeek),
			zap.Ints("days_of_month", req.DaysOfMonth),
			zap.Error(err),
		)
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := config.Cfg.GormDB.Create(&schedule).Error; err != nil {
		apilog.ZapError("create schedule db failed",
			zap.String("user_id", userID),
			zap.String("medicine_id", medicineID),
			zap.String("schedule_id", schedule.ID),
			zap.String("schedule_type", schedule.ScheduleType),
			zap.Ints("days_of_week", req.DaysOfWeek),
			zap.Ints("days_of_month", req.DaysOfMonth),
			zap.Error(err),
		)
		g.JSON(http.StatusInternalServerError, gin.H{"error": "could not create schedule"})
		return
	}
	apilog.ZapInfo("create schedule succeeded",
		zap.String("user_id", userID),
		zap.String("medicine_id", medicineID),
		zap.String("schedule_id", schedule.ID),
		zap.String("schedule_type", schedule.ScheduleType),
		zap.Ints("days_of_week", req.DaysOfWeek),
		zap.Ints("days_of_month", req.DaysOfMonth),
		zap.Int("days_of_week_count", len(schedule.DaysOfWeek)),
		zap.Int("days_of_month_count", len(schedule.DaysOfMonth)),
	)
	g.JSON(http.StatusCreated, gin.H{"schedule": schedule})
}

func UpdateSchedule(g *gin.Context) {
	var existing models.Schedule
	userID := middleware.UserID(g)
	err := config.Cfg.GormDB.First(&existing, "id = ? AND user_id = ?", g.Param("id"), userID).Error
	if err != nil {
		apilog.ZapWarn("update schedule not found",
			zap.String("user_id", userID),
			zap.String("schedule_id", g.Param("id")),
			zap.Error(err),
		)
		g.JSON(http.StatusNotFound, gin.H{"error": "schedule not found"})
		return
	}

	var req scheduleRequest
	if err := g.ShouldBindJSON(&req); err != nil {
		apilog.ZapWarn("update schedule bind failed",
			zap.String("user_id", userID),
			zap.String("medicine_id", existing.MedicineID),
			zap.String("schedule_id", existing.ID),
			zap.Error(err),
		)
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	logScheduleRequest("update schedule request", userID, existing.MedicineID, existing.ID, req)
	schedule, err := buildSchedule(req, userID, existing.MedicineID)
	if err != nil {
		apilog.ZapWarn("update schedule validation failed",
			zap.String("user_id", userID),
			zap.String("medicine_id", existing.MedicineID),
			zap.String("schedule_id", existing.ID),
			zap.String("schedule_type", req.ScheduleType),
			zap.Ints("days_of_week", req.DaysOfWeek),
			zap.Ints("days_of_month", req.DaysOfMonth),
			zap.Error(err),
		)
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	schedule.ID = existing.ID

	err = config.Cfg.GormDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("schedule_id = ?", existing.ID).Delete(&models.DayOfWeek{}).Error; err != nil {
			return err
		}
		if err := tx.Where("schedule_id = ?", existing.ID).Delete(&models.DayOfMonth{}).Error; err != nil {
			return err
		}
		return tx.Save(&schedule).Error
	})
	if err != nil {
		apilog.ZapError("update schedule db failed",
			zap.String("user_id", userID),
			zap.String("medicine_id", existing.MedicineID),
			zap.String("schedule_id", existing.ID),
			zap.String("schedule_type", schedule.ScheduleType),
			zap.Ints("days_of_week", req.DaysOfWeek),
			zap.Ints("days_of_month", req.DaysOfMonth),
			zap.Error(err),
		)
		g.JSON(http.StatusInternalServerError, gin.H{"error": "could not update schedule"})
		return
	}
	apilog.ZapInfo("update schedule succeeded",
		zap.String("user_id", userID),
		zap.String("medicine_id", existing.MedicineID),
		zap.String("schedule_id", existing.ID),
		zap.String("schedule_type", schedule.ScheduleType),
		zap.Ints("days_of_week", req.DaysOfWeek),
		zap.Ints("days_of_month", req.DaysOfMonth),
		zap.Int("days_of_week_count", len(schedule.DaysOfWeek)),
		zap.Int("days_of_month_count", len(schedule.DaysOfMonth)),
	)
	g.JSON(http.StatusOK, gin.H{"schedule": schedule})
}

func DeleteSchedule(g *gin.Context) {
	var schedule models.Schedule
	err := config.Cfg.GormDB.First(&schedule, "id = ? AND user_id = ?", g.Param("id"), middleware.UserID(g)).Error
	if err != nil {
		g.JSON(http.StatusNotFound, gin.H{"error": "schedule not found"})
		return
	}
	err = config.Cfg.GormDB.Transaction(func(tx *gorm.DB) error {
		if err := deleteScheduleDays(tx, []string{schedule.ID}); err != nil {
			return err
		}
		return tx.Delete(&schedule).Error
	})
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete schedule"})
		return
	}
	g.Status(http.StatusNoContent)
}

func deleteMedicineSchedules(tx *gorm.DB, userID, medicineID string) error {
	var scheduleIDs []string
	if err := tx.Model(&models.Schedule{}).
		Where("medicine_id = ? AND user_id = ?", medicineID, userID).
		Pluck("id", &scheduleIDs).Error; err != nil {
		return err
	}
	if len(scheduleIDs) == 0 {
		return nil
	}
	if err := deleteScheduleDays(tx, scheduleIDs); err != nil {
		return err
	}
	return tx.Where("medicine_id = ? AND user_id = ?", medicineID, userID).Delete(&models.Schedule{}).Error
}

func deleteScheduleDays(tx *gorm.DB, scheduleIDs []string) error {
	if err := tx.Where("schedule_id IN ?", scheduleIDs).Delete(&models.DayOfWeek{}).Error; err != nil {
		return err
	}
	return tx.Where("schedule_id IN ?", scheduleIDs).Delete(&models.DayOfMonth{}).Error
}

func buildSchedule(req scheduleRequest, userID, medicineID string) (models.Schedule, error) {
	startDate, err := parseDate(req.StartDate)
	if err != nil {
		return models.Schedule{}, errors.New("start_date must be YYYY-MM-DD")
	}
	var endDate *time.Time
	if req.EndDate != nil && *req.EndDate != "" {
		parsed, err := parseDate(*req.EndDate)
		if err != nil {
			return models.Schedule{}, errors.New("end_date must be YYYY-MM-DD")
		}
		endDate = &parsed
	}
	var cycleStartDate *time.Time
	if req.CycleStartDate != nil && *req.CycleStartDate != "" {
		parsed, err := parseDate(*req.CycleStartDate)
		if err != nil {
			return models.Schedule{}, errors.New("cycle_start_date must be YYYY-MM-DD")
		}
		cycleStartDate = &parsed
	}
	if req.TimeOfDay != nil && !timeOfDayPattern.MatchString(*req.TimeOfDay) {
		return models.Schedule{}, errors.New("time_of_day must be HH:MM")
	}
	if err := validateSchedule(req); err != nil {
		return models.Schedule{}, err
	}

	schedule := models.Schedule{
		ID:                services.NewID(),
		UserID:            userID,
		MedicineID:        medicineID,
		ScheduleType:      req.ScheduleType,
		StartDate:         startDate,
		EndDate:           endDate,
		TimeOfDay:         req.TimeOfDay,
		IntervalHours:     req.IntervalHours,
		IntervalDays:      req.IntervalDays,
		WeeklyInterval:    req.WeeklyInterval,
		MonthlyDay:        req.MonthlyDay,
		CycleActiveDays:   req.CycleActiveDays,
		CycleInactiveDays: req.CycleInactiveDays,
		CycleStartDate:    cycleStartDate,
	}
	for _, day := range req.DaysOfWeek {
		schedule.DaysOfWeek = append(schedule.DaysOfWeek, models.DayOfWeek{ID: services.NewID(), ScheduleID: schedule.ID, Day: day})
	}
	for _, day := range req.DaysOfMonth {
		schedule.DaysOfMonth = append(schedule.DaysOfMonth, models.DayOfMonth{ID: services.NewID(), ScheduleID: schedule.ID, Day: day})
	}
	return schedule, nil
}

func validateSchedule(req scheduleRequest) error {
	switch req.ScheduleType {
	case "daily", "every_other_day":
		return requireTime(req)
	case "weekly":
		if req.WeeklyInterval == nil || *req.WeeklyInterval < 1 {
			return errors.New("weekly_interval must be greater than 0")
		}
		return requireTime(req)
	case "days_of_week":
		if len(req.DaysOfWeek) == 0 {
			return errors.New("days_of_week is required")
		}
		for _, day := range req.DaysOfWeek {
			if day < 0 || day > 6 {
				return errors.New("days_of_week values must be 0-6")
			}
		}
		return requireTime(req)
	case "monthly":
		if req.MonthlyDay == nil || *req.MonthlyDay < 1 || *req.MonthlyDay > 31 {
			return errors.New("monthly_day must be 1-31")
		}
		return requireTime(req)
	case "days_of_month":
		if len(req.DaysOfMonth) == 0 {
			return errors.New("days_of_month is required")
		}
		for _, day := range req.DaysOfMonth {
			if day < 1 || day > 31 {
				return errors.New("days_of_month values must be 1-31")
			}
		}
		return requireTime(req)
	case "cycle":
		if req.CycleActiveDays == nil || req.CycleInactiveDays == nil || *req.CycleActiveDays < 1 || *req.CycleInactiveDays < 1 {
			return errors.New("cycle_active_days and cycle_inactive_days must be greater than 0")
		}
		return requireTime(req)
	case "every_x_hours":
		if req.IntervalHours == nil || *req.IntervalHours < 1 {
			return errors.New("interval_hours must be greater than 0")
		}
		return requireTime(req)
	case "every_x_days":
		if req.IntervalDays == nil || *req.IntervalDays < 1 {
			return errors.New("interval_days must be greater than 0")
		}
		return requireTime(req)
	default:
		return errors.New("invalid schedule_type")
	}
}

func requireTime(req scheduleRequest) error {
	if req.TimeOfDay == nil || *req.TimeOfDay == "" {
		return errors.New("time_of_day is required")
	}
	return nil
}

func parseDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", value)
}

func medicineExists(g *gin.Context, medicineID string) bool {
	var count int64
	config.Cfg.GormDB.Model(&models.Medicine{}).Where("id = ? AND user_id = ?", medicineID, middleware.UserID(g)).Count(&count)
	if count == 0 {
		g.JSON(http.StatusNotFound, gin.H{"error": "medicine not found"})
		return false
	}
	return true
}

func logScheduleRequest(message, userID, medicineID, scheduleID string, req scheduleRequest) {
	apilog.ZapInfo(message,
		zap.String("user_id", userID),
		zap.String("medicine_id", medicineID),
		zap.String("schedule_id", scheduleID),
		zap.String("schedule_type", req.ScheduleType),
		zap.String("start_date", req.StartDate),
		zap.Stringp("end_date", req.EndDate),
		zap.Stringp("time_of_day", req.TimeOfDay),
		zap.Intp("interval_hours", req.IntervalHours),
		zap.Intp("interval_days", req.IntervalDays),
		zap.Intp("weekly_interval", req.WeeklyInterval),
		zap.Intp("monthly_day", req.MonthlyDay),
		zap.Intp("cycle_active_days", req.CycleActiveDays),
		zap.Intp("cycle_inactive_days", req.CycleInactiveDays),
		zap.Stringp("cycle_start_date", req.CycleStartDate),
		zap.Ints("days_of_week", req.DaysOfWeek),
		zap.Ints("days_of_month", req.DaysOfMonth),
	)
}
