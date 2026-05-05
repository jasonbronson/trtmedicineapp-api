package transport

import (
	"github.com/gin-gonic/gin"
)

func HealthCheck(g *gin.Context) {
	g.Writer.Header().Set("Content-Type", "application/json")
	g.Writer.Write([]byte(`{"alive": true}`))
}
