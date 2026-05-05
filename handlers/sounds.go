package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jasonbronson/go-gin-boilerplate/config"
	"github.com/jasonbronson/go-gin-boilerplate/middleware"
	"github.com/jasonbronson/go-gin-boilerplate/models"
	"github.com/jasonbronson/go-gin-boilerplate/services"
)

type soundRequest struct {
	Name     string `json:"name" binding:"required"`
	FileName string `json:"file_name" binding:"required"`
}

func ListSounds(g *gin.Context) {
	var sounds []models.Sound
	config.Cfg.GormDB.Where("user_id = ?", middleware.UserID(g)).Find(&sounds)
	g.JSON(http.StatusOK, gin.H{"sounds": sounds})
}

func CreateSound(g *gin.Context) {
	var req soundRequest
	if err := g.ShouldBindJSON(&req); err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sound := models.Sound{ID: services.NewID(), UserID: middleware.UserID(g), Name: req.Name, FileName: req.FileName}
	if err := config.Cfg.GormDB.Create(&sound).Error; err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"error": "could not create sound"})
		return
	}
	g.JSON(http.StatusCreated, gin.H{"sound": sound})
}

func UpdateSound(g *gin.Context) {
	sound, ok := findSound(g)
	if !ok {
		return
	}
	var req soundRequest
	if err := g.ShouldBindJSON(&req); err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sound.Name = req.Name
	sound.FileName = req.FileName
	if err := config.Cfg.GormDB.Save(&sound).Error; err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"error": "could not update sound"})
		return
	}
	g.JSON(http.StatusOK, gin.H{"sound": sound})
}

func DeleteSound(g *gin.Context) {
	sound, ok := findSound(g)
	if !ok {
		return
	}
	if err := config.Cfg.GormDB.Delete(&sound).Error; err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete sound"})
		return
	}
	g.Status(http.StatusNoContent)
}

func findSound(g *gin.Context) (models.Sound, bool) {
	var sound models.Sound
	if err := config.Cfg.GormDB.First(&sound, "id = ? AND user_id = ?", g.Param("id"), middleware.UserID(g)).Error; err != nil {
		g.JSON(http.StatusNotFound, gin.H{"error": "sound not found"})
		return sound, false
	}
	return sound, true
}
