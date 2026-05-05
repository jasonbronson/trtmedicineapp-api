package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jasonbronson/go-gin-boilerplate/config"
	"github.com/jasonbronson/go-gin-boilerplate/models"
	"github.com/jasonbronson/go-gin-boilerplate/services"
	"gorm.io/gorm"
)

type authRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type googleAuthRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

type changePasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

func Register(g *gin.Context) {
	var req authRequest
	if err := g.ShouldBindJSON(&req); err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	passwordHash, err := services.HashPassword(req.Password)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
		return
	}

	user := models.User{
		ID:            services.NewID(),
		Email:         email,
		PasswordHash:  &passwordHash,
		AuthProvider:  "password",
		EmailVerified: false,
	}
	if err := config.Cfg.GormDB.Create(&user).Error; err != nil {
		g.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		return
	}
	respondWithToken(g, user)
}

func Login(g *gin.Context) {
	var req authRequest
	if err := g.ShouldBindJSON(&req); err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	err := config.Cfg.GormDB.Where("email = ?", strings.ToLower(strings.TrimSpace(req.Email))).First(&user).Error
	if err != nil || user.PasswordHash == nil || !services.CheckPassword(*user.PasswordHash, req.Password) {
		g.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}
	respondWithToken(g, user)
}

func GoogleLogin(g *gin.Context) {
	var req googleAuthRequest
	if err := g.ShouldBindJSON(&req); err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokenInfo, err := services.VerifyGoogleIDToken(req.IDToken)
	if err != nil {
		g.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	email := strings.ToLower(strings.TrimSpace(tokenInfo.Email))
	var user models.User
	err = config.Cfg.GormDB.Where("google_sub = ? OR email = ?", tokenInfo.Subject, email).First(&user).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		g.JSON(http.StatusInternalServerError, gin.H{"error": "could not load user"})
		return
	}

	if err == gorm.ErrRecordNotFound {
		user = models.User{
			ID:            services.NewID(),
			Email:         email,
			GoogleSub:     &tokenInfo.Subject,
			AuthProvider:  "google",
			EmailVerified: true,
		}
		if err := config.Cfg.GormDB.Create(&user).Error; err != nil {
			g.JSON(http.StatusConflict, gin.H{"error": "could not create google user"})
			return
		}
		respondWithToken(g, user)
		return
	}

	updates := map[string]interface{}{
		"google_sub":     tokenInfo.Subject,
		"email_verified": true,
	}
	if user.AuthProvider == "password" {
		updates["auth_provider"] = "password,google"
	}
	if err := config.Cfg.GormDB.Model(&user).Updates(updates).Error; err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"error": "could not link google account"})
		return
	}
	config.Cfg.GormDB.First(&user, "id = ?", user.ID)
	respondWithToken(g, user)
}

func Me(g *gin.Context) {
	var user models.User
	if err := config.Cfg.GormDB.First(&user, "id = ?", g.GetString("user_id")).Error; err != nil {
		g.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	g.JSON(http.StatusOK, gin.H{"user": user})
}

func ChangePassword(g *gin.Context) {
	var req changePasswordRequest
	if err := g.ShouldBindJSON(&req); err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.Cfg.GormDB.First(&user, "id = ?", g.GetString("user_id")).Error; err != nil {
		g.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	passwordHash, err := services.HashPassword(req.NewPassword)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
		return
	}
	updates := map[string]interface{}{"password_hash": passwordHash}
	if user.AuthProvider == "google" {
		updates["auth_provider"] = "password,google"
	}
	if err := config.Cfg.GormDB.Model(&user).Updates(updates).Error; err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"error": "could not update password"})
		return
	}
	g.Status(http.StatusNoContent)
}

func RefreshToken(g *gin.Context) {
	var user models.User
	if err := config.Cfg.GormDB.First(&user, "id = ?", g.GetString("user_id")).Error; err != nil {
		g.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	respondWithToken(g, user)
}

func Logout(g *gin.Context) {
	g.Status(http.StatusNoContent)
}

func respondWithToken(g *gin.Context, user models.User) {
	token, err := services.GenerateToken(user)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	g.JSON(http.StatusOK, gin.H{"access_token": token, "user": user})
}
