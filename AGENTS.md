# API - Go/Gin Backend

## Stack

- **Go 1.22** — module path `github.com/jasonbronson/go-gin-boilerplate`
- **Gin v1.7** — HTTP framework and router
- **GORM v1.25** — ORM with SQLite via `glebarez/sqlite` (pure-Go, no CGO)
- **JWT v4** (`golang-jwt/jwt/v4`) — authentication tokens
- **Zap** (`go.uber.org/zap`) — structured logging
- **New Relic v3** — APM (optional, gated by `NEW_RELIC_ENABLED` env var)
- **Godotenv** — `.env` file loading

## Architecture

```
cmd/api/main.go            → Entry point, starts Gin server
cmd/cron/main.go           → Entry point, runs cron scheduler
transport/routes.go        → All route definitions, middleware wiring
middleware/auth.go         → Bearer JWT auth middleware
handlers/*.go              → HTTP handlers (parse request → call DB/services → JSON response)
services/*.go              → Business logic (JWT, bcrypt, Google OAuth, schedule calculations)
models/models.go           → 8 GORM models (User, Medicine, Schedule, DayOfWeek, DayOfMonth, Sound, ReminderLog, ManualNote)
config/config.go           → Env loading + global DB singleton (config.Cfg.GormDB)
config/migrations.go       → GORM AutoMigrate on startup
config/newrelic.go         → New Relic agent setup
```

## Key Design Patterns

- **Global DB singleton:** `config.Cfg.GormDB` — no dependency injection, handlers access GORM directly
- **Flat handler layer:** No repository/service interface pattern; handlers call GORM queries inline
- **Soft deletes:** `User`, `Medicine`, `Schedule`, and `Sound` use GORM soft delete (has `DeletedAt`). `ReminderLog` and `ManualNote` do NOT — they are hard-deleted.
- **Custom UUIDs:** `services.NewID()` generates deterministic hex IDs (not standard UUID v4)
- **Account deletion:** Uses `Unscoped()` to permanently wipe all user data (bypasses soft delete)
- **Google auth linking:** Users can register with email then link Google. Provider field stores comma-separated values like `"password,google"`
- **Schedule engine:** `services/reminders.go` supports 9 schedule types (daily, every_other_day, every_x_days, every_x_hours, weekly, days_of_week, monthly, days_of_month, cycle)

## Running Locally

```bash
# Docker Compose (hot-reload with CompileDaemon)
make local

# Without Docker (Air hot-reload)
air

# Direct build
make build          # → ./dist/api
make buildcron      # → ./dist/cron

# Tests
make test           # go test -parallel=6 -failfast -cover ./...
```

## Environment Variables (see `.env.example`)

| Variable            | Default                  | Notes                                |
|---------------------|--------------------------|--------------------------------------|
| `PORT`              | `8080`                   |                                      |
| `DATABASE_URL`      | `sqlite3://data/app.db`  |                                      |
| `DB_LOG_MODE`       | `false`                  | Enable GORM SQL logging              |
| `JWT_SECRET`        | *(required)*             | HMAC-SHA256 signing key              |
| `JWT_TOKEN_TTL_HOURS`| `2160` (90 days)        |                                      |
| `GOOGLE_CLIENT_ID`  | *(required for Google)*  |                                      |
| `NEW_RELIC_ENABLED` | `false`                  |                                      |

## Database

- SQLite file at `data/app.db` (gitignored)
- Auto-migrated on startup via `config/migrations.go`
- All GORM models in a single file: `models/models.go`
- No schema versioning or migration scripts — GORM `AutoMigrate` handles schema changes

## Docker

- **Production:** `dockerfile.app` — multi-stage, target `scratch`, copies only binary + CA certs + `medslist.json`
- **Dev:** `docker-compose.yml` — API + cron services with `CompileDaemon` hot-reload
- **CI:** `.github/workflows/docker-build.yml` — builds multi-arch (amd64/arm64), pushes to `ghcr.io/jasonbronson/trtmedicineapp-api`

## API Endpoints Summary

- `GET /`, `GET /healthz` — health checks (public)
- `POST /auth/register`, `POST /auth/login`, `POST /auth/google` — auth (public)
- `POST /auth/refresh`, `POST /auth/logout` — token management
- `GET/DELETE /api/me`, `PUT /api/me/password` — account (protected)
- `GET/POST /api/medicines`, `GET/PUT/DELETE /api/medicines/:id` — medicines CRUD
- `GET /api/medicines/search?q=&limit=` — search from `medslist.json`
- `GET/POST /api/medicines/:id/schedules`, `PUT/DELETE /api/schedules/:id` — schedules
- `GET /api/reminders/due?at=`, `POST /api/reminders/:med_id/taken`, `POST /api/reminders/:med_id/skipped` — reminders
- `GET /api/reminders/history`, `PUT /api/reminders/history/:id/notes` — history
- `GET/POST /api/reminders/manual-notes`, `PUT /api/reminders/manual-notes/:id` — manual notes
- `GET/POST /api/sounds`, `PUT/DELETE /api/sounds/:id` — sounds
- `GET /api/subscription/status`, `POST /api/subscription/apple` — subscriptions

## Notes for AI

- This project has **no test files** despite having a `make test` target
- When adding a new endpoint, follow the existing pattern: define route in `transport/routes.go`, handler in `handlers/`, business logic in `services/` if reusable
- All handlers respond with JSON using Gin's `c.JSON()` or `c.AbortWithStatusJSON()`
- All protected routes go through `middleware.AuthRequired()` which sets `userID` in the Gin context
- The `medslist.json` file is a static search dataset; it's read on startup and searched in-memory
- Do not introduce CGO-dependent packages — the production Docker image is scratch-based