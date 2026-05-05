package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jasonbronson/go-gin-boilerplate/config"
	"github.com/jasonbronson/go-gin-boilerplate/middleware"
	"github.com/jasonbronson/go-gin-boilerplate/models"
)

type appleSubscriptionRequest struct {
	ApplePurchaseReceiptData string  `json:"apple_purchase_receipt_data" binding:"required"`
	AppleOrderID             string  `json:"apple_order_id" binding:"required"`
	PlanID                   string  `json:"plan_id" binding:"required"`
	SubscriptionStatus       string  `json:"subscription_status" binding:"required"`
	CancelDate               *string `json:"cancel_date"`
	SubscriptionExpiresAt    *string `json:"subscription_expires_at"`
}

type subscriptionStatusResponse struct {
	PlanID                *string    `json:"plan_id"`
	SubscriptionStatus    string     `json:"subscription_status"`
	CancelDate            *time.Time `json:"cancel_date,omitempty"`
	SubscriptionExpiresAt *time.Time `json:"subscription_expires_at,omitempty"`
}

func GetSubscriptionStatus(g *gin.Context) {
	var user models.User
	if err := config.Cfg.GormDB.First(&user, "id = ?", middleware.UserID(g)).Error; err != nil {
		g.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	g.JSON(http.StatusOK, gin.H{"subscription": subscriptionResponse(user)})
}

func UpdateAppleSubscription(g *gin.Context) {
	var req appleSubscriptionRequest
	if err := g.ShouldBindJSON(&req); err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := strings.ToLower(strings.TrimSpace(req.SubscriptionStatus))
	switch status {
	case "active", "inactive", "expired", "canceled", "revoked":
	default:
		g.JSON(http.StatusBadRequest, gin.H{"error": "subscription_status is invalid"})
		return
	}

	cancelDate, err := parseOptionalRFC3339(req.CancelDate)
	if err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": "cancel_date must be RFC3339"})
		return
	}
	expiresAt, err := parseOptionalRFC3339(req.SubscriptionExpiresAt)
	if err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": "subscription_expires_at must be RFC3339"})
		return
	}

	updates := map[string]interface{}{
		"apple_purchase_receipt_data": strings.TrimSpace(req.ApplePurchaseReceiptData),
		"apple_order_id":              strings.TrimSpace(req.AppleOrderID),
		"plan_id":                     strings.TrimSpace(req.PlanID),
		"subscription_status":         status,
		"cancel_date":                 cancelDate,
		"subscription_expires_at":     expiresAt,
	}

	if err := config.Cfg.GormDB.Model(&models.User{}).Where("id = ?", middleware.UserID(g)).Updates(updates).Error; err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"error": "could not update subscription"})
		return
	}

	var user models.User
	if err := config.Cfg.GormDB.First(&user, "id = ?", middleware.UserID(g)).Error; err != nil {
		g.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	g.JSON(http.StatusOK, gin.H{"subscription": subscriptionResponse(user)})
}

func subscriptionResponse(user models.User) subscriptionStatusResponse {
	status := user.SubscriptionStatus
	if status == "" {
		status = "inactive"
	}
	return subscriptionStatusResponse{
		PlanID:                user.PlanID,
		SubscriptionStatus:    status,
		CancelDate:            user.CancelDate,
		SubscriptionExpiresAt: user.SubscriptionExpiresAt,
	}
}

func parseOptionalRFC3339(value *string) (*time.Time, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(*value))
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
