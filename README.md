# StudyBuddy Backend (Go)

This is a simple Go backend providing RESTful API endpoints for the StudyBuddy app. It's intentionally minimal and uses an in-memory store (no external DB) so it is easy to run locally for development.

## Features
- REST API for Tasks, Courses, Events
- User registration & login with JWT authentication
- In-memory store seeded with sample data matching the app's reference data
- CORS middleware, logging

## Quick Start

1. Install Go 1.20+.
2. From this directory, run:

```bash
# Get deps
go mod tidy
# Start server
go run .
```

3. Server will run at http://localhost:8080 by default.

## Quick examples

Login with seeded user:

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password"}'
```

Get tasks (no auth required):

```bash
curl http://localhost:8080/api/tasks
```

Create task (auth required):

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login -H "Content-Type: application/json" -d '{"email":"test@example.com","password":"password"}' | jq -r .token)
curl -X POST http://localhost:8080/api/tasks -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"title":"New Task","description":"example","courseId":"1","dueDate":"2025-12-01","dueTime":"12:00","completed":false,"hasReminder":true}'
```


## Auth

- Register: POST /api/auth/register
  - body: {"name":"...","email":"...","password":"..."}
- Login: POST /api/auth/login
  - body: {"email":"...","password":"..."}
  - returns {"token":"..."}

For endpoints that require authentication, include the header:

```
Authorization: Bearer <token>
```

Seeded test user (for quick development):
- email: `test@example.com`
- password: `password`

## API Endpoints
- GET /api/tasks
- POST /api/tasks (auth)
- GET /api/tasks/:id
- PUT /api/tasks/:id (auth)
- DELETE /api/tasks/:id (auth)

- GET /api/courses
- POST /api/courses (auth)

- GET /api/events
- POST /api/events (auth)

## Notes
- This backend uses an in-memory store; restart will lose data.
- To persist or deploy, replace the store with a DB (Postgres, SQLite) and add migrations.
- JWT secret: set `JWT_SECRET` env var. Default is `dev-secret`.
- Port: controlled by `PORT` env var (default 8080).
- MongoDB: set `MONGO_URI` env var to a MongoDB URI (e.g., mongodb://localhost:27017). If set, the app will use MongoDB for persistence; otherwise it defaults to an in-memory store.

IMPORTANT: Never commit credentials directly into source. Set them via environment variables or a secret manager. Below is a sample **run-only** example where you might have been provided a URI (don't commit this into code):

```bash
export MONGO_URI="mongodb+srv://<username>:<password>@cluster0.tqyenfu.mongodb.net/?appName=Cluster0"
export JWT_SECRET="supersecret"
go run .
```

If you provide a URI (like the sample above) the app will use your MongoDB cluster and seed collections on startup, but avoid keeping secrets in your repository or code files.

Using MongoDB (example):

```bash
# Start a local MongoDB or use a cloud URI
export MONGO_URI="mongodb://localhost:27017"
export JWT_SECRET="supersecret"
go run .
```

The app will use database `studybuddy` and seed collections on startup (tasks, courses, events, users).

Dotenv (.env) support
---------------------
You can also create a `.env` file in the `StudyBuddy_Backend` folder with default values (see `.env.example` for the format).
System environment variables (e.g. from your shell or CI) will always override values in `.env`.

Example `.env` file:
```
MONGO_URI="mongodb://localhost:27017"
JWT_SECRET="dev-secret"
PORT="8080"
```

## Next steps (optional)
- Add persistent storage and migrations
- Add tests
- Add production-ready logging & configuration
