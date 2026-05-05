package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/jasonbronson/go-gin-boilerplate/config"
	apilog "github.com/jasonbronson/go-gin-boilerplate/library/log"
	"github.com/jasonbronson/go-gin-boilerplate/middleware"
	"github.com/jasonbronson/go-gin-boilerplate/models"
	"github.com/jasonbronson/go-gin-boilerplate/services"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type medicineRequest struct {
	Name       string  `json:"name" binding:"required"`
	DoseAmount string  `json:"dose_amount"`
	DoseUnit   string  `json:"dose_unit"`
	Notes      string  `json:"notes"`
	SoundID    *string `json:"sound_id"`
	Active     *bool   `json:"active"`
}

type medListFile struct {
	DisplayTermsList struct {
		Term []string `json:"term"`
	} `json:"displayTermsList"`
}

var (
	medListOnce sync.Once
	medList     []string
	medListErr  error
)

func ListMedicines(g *gin.Context) {
	var medicines []models.Medicine
	config.Cfg.GormDB.Preload("Schedules.DaysOfWeek").Preload("Schedules.DaysOfMonth").Where("user_id = ?", middleware.UserID(g)).Find(&medicines)
	g.JSON(http.StatusOK, gin.H{"medicines": medicines})
}

func SearchMedicineList(g *gin.Context) {
	query := strings.TrimSpace(g.Query("q"))
	if query == "" {
		g.JSON(http.StatusBadRequest, gin.H{"error": "q is required"})
		return
	}

	limit := 25
	if rawLimit := g.Query("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil || parsedLimit < 1 {
			g.JSON(http.StatusBadRequest, gin.H{"error": "limit must be a positive number"})
			return
		}
		if parsedLimit < limit {
			limit = parsedLimit
		}
	}

	meds, err := loadMedList()
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"error": "could not load medicine list"})
		return
	}

	normalizedQuery := strings.ToLower(query)
	matches := make([]string, 0, limit)
	for _, med := range meds {
		if strings.Contains(strings.ToLower(med), normalizedQuery) {
			matches = append(matches, med)
			if len(matches) == limit {
				break
			}
		}
	}

	g.JSON(http.StatusOK, gin.H{"matches": matches})
}

func CreateMedicine(g *gin.Context) {
	var req medicineRequest
	if err := g.ShouldBindJSON(&req); err != nil {
		apilog.ZapWarn("create medicine bind failed",
			zap.String("user_id", middleware.UserID(g)),
			zap.Error(err),
		)
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	apilog.ZapInfo("create medicine request",
		zap.String("user_id", middleware.UserID(g)),
		zap.String("name", req.Name),
		zap.String("dose_amount", req.DoseAmount),
		zap.String("dose_unit", req.DoseUnit),
		zap.String("notes", req.Notes),
		zap.Stringp("sound_id", req.SoundID),
		zap.Boolp("active", req.Active),
	)
	active := true
	if req.Active != nil {
		active = *req.Active
	}
	medicine := models.Medicine{
		ID:         services.NewID(),
		UserID:     middleware.UserID(g),
		Name:       req.Name,
		DoseAmount: req.DoseAmount,
		DoseUnit:   req.DoseUnit,
		Notes:      req.Notes,
		SoundID:    req.SoundID,
		Active:     active,
	}
	if err := config.Cfg.GormDB.Create(&medicine).Error; err != nil {
		apilog.ZapError("create medicine db failed",
			zap.String("user_id", medicine.UserID),
			zap.String("medicine_id", medicine.ID),
			zap.String("name", medicine.Name),
			zap.Error(err),
		)
		g.JSON(http.StatusInternalServerError, gin.H{"error": "could not create medicine"})
		return
	}
	apilog.ZapInfo("create medicine succeeded",
		zap.String("user_id", medicine.UserID),
		zap.String("medicine_id", medicine.ID),
		zap.String("name", medicine.Name),
		zap.Bool("active", medicine.Active),
	)
	g.JSON(http.StatusCreated, gin.H{"medicine": medicine})
}

func GetMedicine(g *gin.Context) {
	medicine, ok := findMedicine(g)
	if !ok {
		return
	}
	g.JSON(http.StatusOK, gin.H{"medicine": medicine})
}

func UpdateMedicine(g *gin.Context) {
	medicine, ok := findMedicine(g)
	if !ok {
		return
	}
	var req medicineRequest
	if err := g.ShouldBindJSON(&req); err != nil {
		apilog.ZapWarn("update medicine bind failed",
			zap.String("user_id", middleware.UserID(g)),
			zap.String("medicine_id", medicine.ID),
			zap.Error(err),
		)
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	apilog.ZapInfo("update medicine request",
		zap.String("user_id", middleware.UserID(g)),
		zap.String("medicine_id", medicine.ID),
		zap.String("name", req.Name),
		zap.String("dose_amount", req.DoseAmount),
		zap.String("dose_unit", req.DoseUnit),
		zap.String("notes", req.Notes),
		zap.Stringp("sound_id", req.SoundID),
		zap.Boolp("active", req.Active),
	)
	medicine.Name = req.Name
	medicine.DoseAmount = req.DoseAmount
	medicine.DoseUnit = req.DoseUnit
	medicine.Notes = req.Notes
	medicine.SoundID = req.SoundID
	if req.Active != nil {
		medicine.Active = *req.Active
	}
	if err := config.Cfg.GormDB.Save(&medicine).Error; err != nil {
		apilog.ZapError("update medicine db failed",
			zap.String("user_id", medicine.UserID),
			zap.String("medicine_id", medicine.ID),
			zap.String("name", medicine.Name),
			zap.Error(err),
		)
		g.JSON(http.StatusInternalServerError, gin.H{"error": "could not update medicine"})
		return
	}
	apilog.ZapInfo("update medicine succeeded",
		zap.String("user_id", medicine.UserID),
		zap.String("medicine_id", medicine.ID),
		zap.String("name", medicine.Name),
		zap.Bool("active", medicine.Active),
	)
	g.JSON(http.StatusOK, gin.H{"medicine": medicine})
}

func DeleteMedicine(g *gin.Context) {
	medicine, ok := findMedicine(g)
	if !ok {
		return
	}
	err := config.Cfg.GormDB.Transaction(func(tx *gorm.DB) error {
		if err := deleteMedicineSchedules(tx, middleware.UserID(g), medicine.ID); err != nil {
			return err
		}
		return tx.Delete(&medicine).Error
	})
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete medicine"})
		return
	}
	g.Status(http.StatusNoContent)
}

func findMedicine(g *gin.Context) (models.Medicine, bool) {
	var medicine models.Medicine
	err := config.Cfg.GormDB.Preload("Schedules.DaysOfWeek").Preload("Schedules.DaysOfMonth").First(&medicine, "id = ? AND user_id = ?", g.Param("id"), middleware.UserID(g)).Error
	if err != nil {
		g.JSON(http.StatusNotFound, gin.H{"error": "medicine not found"})
		return medicine, false
	}
	return medicine, true
}

func loadMedList() ([]string, error) {
	medListOnce.Do(func() {
		var raw []byte
		for _, path := range medListPaths() {
			raw, medListErr = os.ReadFile(path)
			if medListErr == nil {
				break
			}
		}
		if medListErr != nil {
			return
		}

		var file medListFile
		if err := json.Unmarshal(raw, &file); err != nil {
			medListErr = err
			return
		}
		medList = file.DisplayTermsList.Term
	})

	return medList, medListErr
}

func medListPaths() []string {
	return []string{
		"medlist.json",
		"medslist.json",
		filepath.Join("api", "medlist.json"),
		filepath.Join("api", "medslist.json"),
	}
}
