package config

import (
	"github.com/jasonbronson/go-gin-boilerplate/library/log"
	"github.com/jasonbronson/go-gin-boilerplate/models"
)

func AutoMigrate() {
	if err := Cfg.GormDB.AutoMigrate(
		&models.User{},
		&models.Medicine{},
		&models.Schedule{},
		&models.DayOfWeek{},
		&models.DayOfMonth{},
		&models.Sound{},
		&models.ReminderLog{},
		&models.ManualNote{},
	); err != nil {
		log.Fatalf("could not migrate database: %v", err)
	}
}
