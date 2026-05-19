package transport

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jasonbronson/go-gin-boilerplate/handlers"
	"github.com/jasonbronson/go-gin-boilerplate/middleware"
	"github.com/newrelic/go-agent/v3/integrations/nrgin"
	"github.com/newrelic/go-agent/v3/newrelic"
	requestid "github.com/sumit-tembe/gin-requestid"
)

// Router func
func Router(newRelicApp *newrelic.Application) http.Handler {

	corsConfig := cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}
	router := gin.Default()
	router.RedirectTrailingSlash = true

	router.Use(nrgin.Middleware(newRelicApp))
	router.Use(cors.New(corsConfig))
	router.Use(requestid.RequestID(nil))

	// if os.Getenv("ENVIRONMENT") == "load" {
	// 	pprof.Register(router, "debug")
	// }

	router.GET("/", HealthCheck)
	router.GET("/healthz", HealthCheck)

	auth := router.Group("/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
		auth.POST("/google", handlers.GoogleLogin)
		auth.POST("/refresh", middleware.AuthRequired(), handlers.RefreshToken)
		auth.POST("/logout", middleware.AuthRequired(), handlers.Logout)
	}

	api := router.Group("/api")
	api.Use(middleware.AuthRequired())
	{
		api.GET("/me", handlers.Me)
		api.DELETE("/me", handlers.DeleteAccount)
		api.PUT("/me/password", handlers.ChangePassword)
		api.GET("/subscription/status", handlers.GetSubscriptionStatus)
		api.POST("/subscription/apple", handlers.UpdateAppleSubscription)

		api.GET("/medicines", handlers.ListMedicines)
		api.POST("/medicines", handlers.CreateMedicine)
		api.GET("/medicines/search", handlers.SearchMedicineList)
		api.GET("/medicines/:id", handlers.GetMedicine)
		api.PUT("/medicines/:id", handlers.UpdateMedicine)
		api.DELETE("/medicines/:id", handlers.DeleteMedicine)

		api.GET("/medicines/:id/schedules", handlers.ListSchedules)
		api.POST("/medicines/:id/schedules", handlers.CreateSchedule)
		api.PUT("/schedules/:id", handlers.UpdateSchedule)
		api.DELETE("/schedules/:id", handlers.DeleteSchedule)

		api.GET("/sounds", handlers.ListSounds)
		api.POST("/sounds", handlers.CreateSound)
		api.PUT("/sounds/:id", handlers.UpdateSound)
		api.DELETE("/sounds/:id", handlers.DeleteSound)

		api.GET("/reminders/due", handlers.DueReminders)
		api.POST("/reminders/:medicine_id/taken", handlers.MarkReminderTaken)
		api.POST("/reminders/:medicine_id/skipped", handlers.MarkReminderSkipped)
		api.GET("/reminders/history", handlers.ReminderHistory)
		api.PUT("/reminders/history/:id/notes", handlers.UpdateReminderHistoryNotes)
		api.GET("/reminders/manual-notes", handlers.ListManualNotes)
		api.POST("/reminders/manual-notes", handlers.CreateManualNote)
		api.PUT("/reminders/manual-notes/:id", handlers.UpdateManualNote)
	}

	//Performance verify key on load forge
	loaderVerification := os.Getenv("LOAD_FORGE")
	if loaderVerification != "" {
		router.GET(fmt.Sprintf("/%v", loaderVerification), func(g *gin.Context) {
			g.Writer.WriteHeader(http.StatusOK)
			g.Writer.Write([]byte(loaderVerification))
		})
	}

	return router
}
