package config

import (
	"database/sql"
	"os"
	"path/filepath"
	"strconv"
	"time"

	logo "log"

	"github.com/jasonbronson/go-gin-boilerplate/library/log"
	"github.com/joho/godotenv"
	_ "github.com/newrelic/go-agent/v3/integrations/nrsqlite3"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/xo/dburl"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	Cfg    = &Config{}
	Driver = "nrsqlite3"
)

type Config struct {
	Port               int
	DatabaseURL        string
	DBLogMode          bool
	JWTSecret          string
	JWTIssuer          string
	JWTAudience        string
	JWTTokenTTL        time.Duration
	GoogleClientID     string
	SQLDB              *sql.DB
	GormDB             *gorm.DB
	NewRelicEnabled    bool
	NewRelicLicenseKey string
	NewRelicAppName    string
	NewRelicApp        *newrelic.Application
}

func init() {
	initEnv()
	initDB()
}

func initEnv() {
	godotenv.Load()
	Cfg.Port, _ = strconv.Atoi(os.Getenv("PORT"))
	if Cfg.Port == 0 {
		Cfg.Port = 8080
	}
	Cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if Cfg.DatabaseURL == "" {
		Cfg.DatabaseURL = "sqlite3://data/app.db"
	}
	Cfg.DBLogMode, _ = strconv.ParseBool(os.Getenv("DB_LOG_MODE"))
	Cfg.JWTSecret = os.Getenv("JWT_SECRET")
	Cfg.JWTIssuer = os.Getenv("JWT_ISSUER")
	Cfg.JWTAudience = os.Getenv("JWT_AUDIENCE")
	jwtTokenTTLHours, _ := strconv.Atoi(os.Getenv("JWT_TOKEN_TTL_HOURS"))
	if jwtTokenTTLHours == 0 {
		jwtTokenTTLHours = 24 * 90
	}
	Cfg.JWTTokenTTL = time.Duration(jwtTokenTTLHours) * time.Hour
	Cfg.GoogleClientID = os.Getenv("GOOGLE_CLIENT_ID")
	Cfg.NewRelicEnabled, _ = strconv.ParseBool(os.Getenv("NEW_RELIC_ENABLED"))
	Cfg.NewRelicLicenseKey = os.Getenv("NEW_RELIC_LICENSE_KEY")
	Cfg.NewRelicAppName = os.Getenv("NEW_RELIC_APP_NAME")

}

func initDB() {

	var err error
	u, err := dburl.Parse(Cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	dbname := u.DSN
	if dbname == "" {
		log.Fatal("database not found or empty env var")
	}
	if dir := filepath.Dir(dbname); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("could not create database directory: %v", err)
		}
	}
	dbdialect := sqlite.Open(dbname)

	newLogger := logger.New(
		logo.New(logo.Writer(), "GORMDB:", logo.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second * 2, // Slow SQL threshold
			LogLevel:                  logger.Error,    // Log level
			IgnoreRecordNotFoundError: true,            // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,            // Enable color
		},
	)

	gconfig := &gorm.Config{}
	if Cfg.GormDB, err = gorm.Open(dbdialect, gconfig); err != nil {
		log.Fatalf("could not initialize gorm: %v", err)
	}

	//Debug SQL logs?
	if Cfg.DBLogMode {
		Cfg.GormDB.Logger = newLogger
	}

	log.Println("Success connecting to database")
}
