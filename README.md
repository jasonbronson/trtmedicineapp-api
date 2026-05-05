# TRT Med Reminders API
Gin, Go, SQLite, JWT auth, Google login, New Relic, Zap logger, and Cron.

## Quick Start
- To run project you can use make command ```make local```
- if you do not have make installed just use docker-compose ```docker-compose up```

## Configuration

- copy .env.example to .env ```cp .env.example .env```
- edit .env as needed

Required auth config:

```
DATABASE_URL=sqlite3://data/app.db
JWT_SECRET=change-me
JWT_ISSUER=trt-medicare-api
JWT_AUDIENCE=trt-medicare-ios
GOOGLE_CLIENT_ID=your-ios-google-client-id
```
