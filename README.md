# User Management Lifecycle Application

Full-stack user management: create, view, update, activate, deactivate, search, filter and delete users.

- **Frontend:** React + TypeScript (Vite), in `frontend/`
- **Backend:** Go REST API (Handler → Service → Repository), in `backend/`
- **Database:** MongoDB via the official Go driver (`users` collection, unique index on `email`)

## Run with Docker Compose

```bash
cp .env.example .env
docker compose up --build
```

- Frontend: http://localhost:3000
- Backend: http://localhost:8080 (health check: `/healthz`)

## Configuration

All configuration comes from environment variables; nothing is hardcoded.

| Variable | Used by | Description |
|---|---|---|
| `MONGO_URI` | backend | MongoDB connection string (required) |
| `MONGO_DB` | backend | Database name (required) |
| `CORS_ALLOWED_ORIGIN` | backend | The only origin allowed by CORS (required) |
| `PORT` | backend | Listen port (default `8080`) |
| `MAX_BODY_BYTES` | backend | Max request body size (default `1048576`) |
| `LOG_LEVEL` | backend | `debug`, `info`, `warn`, `error` (default `info`) |
| `VITE_API_BASE_URL` | frontend | Backend URL as seen from the browser (build time) |

Examples: `.env.example` (Compose), `backend/.env.example`, `frontend/.env.example`.

## Run locally without Docker

```bash
# backend (needs a reachable MongoDB)
cd backend
export MONGO_URI=mongodb://localhost:27017 MONGO_DB=usermanagement CORS_ALLOWED_ORIGIN=http://localhost:3000
go run ./cmd/server

# frontend
cd frontend
npm install
VITE_API_BASE_URL=http://localhost:8080 npm run dev
```

## API

All responses use the envelope `{ "success": bool, "data": ..., "error": { "code", "message", "details" } }`.

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/users` | Create a user (`201`, `400`, `409`) |
| GET | `/api/v1/users?page&pageSize&search&status` | List with pagination, search (name/email), status filter |
| GET | `/api/v1/users/stats` | Total, active and inactive counts (dashboard) |
| GET | `/api/v1/users/{id}` | Get a user (`400` malformed id, `404` missing) |
| PUT | `/api/v1/users/{id}` | Update a user (`400`, `404`, `409`) |
| PATCH | `/api/v1/users/{id}/status` | Body `{"status":"ACTIVE"|"INACTIVE"}` |
| DELETE | `/api/v1/users/{id}` | Permanently delete (`404` if missing) |

When MongoDB is unavailable the API returns `503` with a generic message (no connection string, host or stack trace).

## Testing

```bash
# backend unit and API tests
cd backend && go test -count=1 ./...

# backend repository tests against a real MongoDB (skipped when MONGO_TEST_URI is unset)
cd backend && MONGO_TEST_URI=mongodb://localhost:27017 go test -count=1 ./internal/repository/...

# frontend component tests and build
cd frontend && npm install && npx vitest run && npm run build
```

To get a MongoDB for the repository tests: `docker compose up -d mongo`, then publish its port or run
`docker run -d -p 27017:27017 mongo:7`. Tests use a throwaway database that is dropped afterwards.
