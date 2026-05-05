package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jasonbronson/go-gin-boilerplate/config"
	"github.com/jasonbronson/go-gin-boilerplate/middleware"
	"github.com/jasonbronson/go-gin-boilerplate/models"
	"github.com/jasonbronson/go-gin-boilerplate/services"
)

type reminderActionRequest struct {
	ScheduleID string `json:"schedule_id" binding:"required"`
	DueAt      string `json:"due_at" binding:"required"`
}

type reminderNoteRequest struct {
	Notes string `json:"notes"`
}

type manualNoteRequest struct {
	MedicineID *string `json:"medicine_id"`
	NoteDate   string  `json:"note_date" binding:"required"`
	Notes      string  `json:"notes" binding:"required"`
}

func DueReminders(g *gin.Context) {
	now := time.Now()
	if at := g.Query("at"); at != "" {
		parsed, err := time.Parse(time.RFC3339, at)
		if err != nil {
			g.JSON(http.StatusBadRequest, gin.H{"error": "at must be RFC3339"})
			return
		}
		now = parsed
	}

	var medicines []models.Medicine
	config.Cfg.GormDB.Preload("Schedules.DaysOfWeek").Preload("Schedules.DaysOfMonth").
		Where("user_id = ? AND active = ?", middleware.UserID(g), true).
		Find(&medicines)

	due := []services.DueReminder{}
	for _, medicine := range medicines {
		for _, schedule := range medicine.Schedules {
			if !services.IsScheduleDue(schedule, now) {
				continue
			}
			dueAt := services.DueAt(schedule, now)
			var count int64
			config.Cfg.GormDB.Model(&models.ReminderLog{}).
				Where("user_id = ? AND schedule_id = ? AND due_at = ?", middleware.UserID(g), schedule.ID, dueAt).
				Count(&count)
			if count == 0 {
				due = append(due, services.DueReminder{Medicine: medicine, Schedule: schedule, DueAt: dueAt})
			}
		}
	}
	g.JSON(http.StatusOK, gin.H{"reminders": due})
}

func MarkReminderTaken(g *gin.Context) {
	createReminderLog(g, "taken")
}

func MarkReminderSkipped(g *gin.Context) {
	createReminderLog(g, "skipped")
}

func ReminderHistory(g *gin.Context) {
	var logs []models.ReminderLog
	config.Cfg.GormDB.Where("user_id = ?", middleware.UserID(g)).Order("due_at desc").Limit(200).Find(&logs)
	g.JSON(http.StatusOK, gin.H{"history": logs})
}

func UpdateReminderHistoryNotes(g *gin.Context) {
	var req reminderNoteRequest
	if err := g.ShouldBindJSON(&req); err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var log models.ReminderLog
	if err := config.Cfg.GormDB.First(&log, "id = ? AND user_id = ?", g.Param("id"), middleware.UserID(g)).Error; err != nil {
		g.JSON(http.StatusNotFound, gin.H{"error": "history entry not found"})
		return
	}

	log.Notes = req.Notes
	if err := config.Cfg.GormDB.Save(&log).Error; err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"error": "could not save notes"})
		return
	}
	g.JSON(http.StatusOK, gin.H{"reminder_log": log})
}

func ListManualNotes(g *gin.Context) {
	var notes []models.ManualNote
	config.Cfg.GormDB.Where("user_id = ?", middleware.UserID(g)).Order("note_date desc, created_at desc").Limit(200).Find(&notes)
	g.JSON(http.StatusOK, gin.H{"notes": notes})
}

func CreateManualNote(g *gin.Context) {
	var req manualNoteRequest
	if err := g.ShouldBindJSON(&req); err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	noteDate, ok := parseNoteDate(g, req.NoteDate)
	if !ok {
		return
	}
	if !validateOptionalMedicine(g, req.MedicineID) {
		return
	}
	note := models.ManualNote{
		ID:         services.NewID(),
		UserID:     middleware.UserID(g),
		MedicineID: req.MedicineID,
		NoteDate:   noteDate,
		Notes:      req.Notes,
	}
	if err := config.Cfg.GormDB.Create(&note).Error; err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"error": "could not save note"})
		return
	}
	g.JSON(http.StatusCreated, gin.H{"note": note})
}

func UpdateManualNote(g *gin.Context) {
	var req manualNoteRequest
	if err := g.ShouldBindJSON(&req); err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var note models.ManualNote
	if err := config.Cfg.GormDB.First(&note, "id = ? AND user_id = ?", g.Param("id"), middleware.UserID(g)).Error; err != nil {
		g.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}
	noteDate, ok := parseNoteDate(g, req.NoteDate)
	if !ok {
		return
	}
	if !validateOptionalMedicine(g, req.MedicineID) {
		return
	}
	note.MedicineID = req.MedicineID
	note.NoteDate = noteDate
	note.Notes = req.Notes
	if err := config.Cfg.GormDB.Save(&note).Error; err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"error": "could not save note"})
		return
	}
	g.JSON(http.StatusOK, gin.H{"note": note})
}

func createReminderLog(g *gin.Context, action string) {
	var req reminderActionRequest
	if err := g.ShouldBindJSON(&req); err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dueAt, err := time.Parse(time.RFC3339, req.DueAt)
	if err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": "due_at must be RFC3339"})
		return
	}
	var schedule models.Schedule
	if err := config.Cfg.GormDB.First(&schedule, "id = ? AND medicine_id = ? AND user_id = ?", req.ScheduleID, g.Param("medicine_id"), middleware.UserID(g)).Error; err != nil {
		g.JSON(http.StatusNotFound, gin.H{"error": "schedule not found"})
		return
	}
	var medicine models.Medicine
	if err := config.Cfg.GormDB.First(&medicine, "id = ? AND user_id = ?", g.Param("medicine_id"), middleware.UserID(g)).Error; err != nil {
		g.JSON(http.StatusNotFound, gin.H{"error": "medicine not found"})
		return
	}
	log := models.ReminderLog{
		ID:         services.NewID(),
		UserID:     middleware.UserID(g),
		MedicineID: medicine.ID,
		ScheduleID: req.ScheduleID,
		DoseAmount: medicine.DoseAmount,
		DoseUnit:   medicine.DoseUnit,
		DueAt:      dueAt,
		Action:     action,
	}
	if err := config.Cfg.GormDB.Create(&log).Error; err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"error": "could not save reminder log"})
		return
	}
	g.JSON(http.StatusCreated, gin.H{"reminder_log": log})
}

func parseNoteDate(g *gin.Context, value string) (time.Time, bool) {
	noteDate, err := time.Parse("2006-01-02", value)
	if err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": "note_date must be YYYY-MM-DD"})
		return time.Time{}, false
	}
	return noteDate, true
}

func validateOptionalMedicine(g *gin.Context, medicineID *string) bool {
	if medicineID == nil || *medicineID == "" {
		return true
	}
	var count int64
	config.Cfg.GormDB.Model(&models.Medicine{}).Where("id = ? AND user_id = ?", *medicineID, middleware.UserID(g)).Count(&count)
	if count == 0 {
		g.JSON(http.StatusNotFound, gin.H{"error": "medicine not found"})
		return false
	}
	return true
}
