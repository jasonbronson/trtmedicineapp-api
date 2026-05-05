package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jasonbronson/go-gin-boilerplate/services"
)

func AuthRequired() gin.HandlerFunc {
	return func(g *gin.Context) {
		authHeader := g.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			g.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		claims, err := services.ParseToken(strings.TrimPrefix(authHeader, "Bearer "))
		if err != nil {
			g.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		g.Set("user_id", claims.UserID)
		g.Set("email", claims.Email)
		g.Next()
	}
}

func UserID(g *gin.Context) string {
	userID, _ := g.Get("user_id")
	id, _ := userID.(string)
	return id
}
