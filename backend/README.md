# Backend (Go): AI Payment Support System

The REST API behind the React frontend, written in Go.

**At a glance:** Go 1.22+ standard library router · MySQL via `database/sql` · JWT auth · 24 API endpoints · 26 tests, run with the race detector

## Why it is built this way

| Decision | Reason |
|---|---|
| **Money as integer paise** (`money.Amount`) | `0.10 + 0.20` is always exactly `0.30`. No floating-point errors in payments. |
| **Compare-and-set status updates** (`UPDATE ... WHERE id = ? AND status = ?`) | If two admins update the same refund at once, only one wins and the other gets `409 Conflict`. |
| **`UNIQUE(transaction_id)` on refunds** | Ten simultaneous refund requests for one payment create exactly one refund (tested). |
| **State machines for statuses** | Refund: `REQUESTED → PROCESSING → COMPLETED`. Ticket: `OPEN → IN_PROGRESS → CLOSED`. Transaction: only `PENDING` can change. Every other move is rejected. |
| **Salted PBKDF2-SHA256 passwords** (29,000 rounds, constant-time compare) | Passwords are never stored in plain text and are slow to brute-force. |
| **JWT auth** (HS256, expiry required, `alg: none` rejected) | Stateless login for the React app. |
| **Plain SQL, no ORM** | Every query is visible and parameterised (no SQL injection). Dashboard counts use SQL `COUNT`/`SUM`. |
| **`Store` interface** | MySQL in production, in-memory for tests and for running without a database. |
| **PDF receipts with the standard library** | No extra dependency; the PDF is built in memory. |
| **Graceful shutdown, server timeouts, panic recovery, JSON request logs** | Production basics. |

## Project structure

```
backend/
├── cmd/server/          main.go: config, DB connection, HTTP server, graceful shutdown
└── internal/
    ├── api/             routes, auth/admin middleware, CORS, logging, handlers
    ├── service/         business rules (validation, refunds, tickets, dashboards)
    ├── domain/          models, enums, status state machines
    ├── store/           Store interface + MySQL and in-memory implementations
    ├── auth/            PBKDF2 password hashing, JWT
    ├── money/           paise-based Amount type
    ├── receipt/         PDF receipt builder
    └── config/          environment variables and .env loading
```

Request flow: `api` (HTTP) → `service` (rules) → `store` (MySQL).

## Run it

Requires Go 1.22 or newer.

```bash
cd backend
go mod tidy                 # downloads dependencies and creates go.sum
cp .env.example .env        # then edit the values
go run ./cmd/server         # starts on http://localhost:8080
```

- **With MySQL:** set `DATABASE_URL` (e.g. `mysql://user:pass@localhost:3306/payment_support_db`). Missing tables are created on start.
- **Without a database:** leave `DATABASE_URL` empty and data is kept in memory (lost on restart). Useful for a quick demo.
- **Admin account:** set `ADMIN_EMAIL` and `ADMIN_PASSWORD` and that user is created or promoted to admin on start.

To use it from the frontend, set `VITE_API_BASE_URL=http://localhost:8080` in `frontend/.env`.

## Environment variables

| Variable | Required | Default |
|---|---|---|
| `JWT_SECRET_KEY` | yes (16+ characters) | |
| `DATABASE_URL` | no | in-memory store |
| `PORT` | no | `8080` |
| `ACCESS_TOKEN_EXPIRE_MINUTES` | no | `60` |
| `CORS_ORIGINS` | no | `http://localhost:5173,https://ai-payment-support-system.vercel.app` |
| `ADMIN_EMAIL`, `ADMIN_PASSWORD` | no | |

## Tests

```bash
go test -race ./...
```

26 tests cover login and token checks (expired, wrong secret, `alg: none`), password hashing, money parsing, every status state machine, ownership (a user can't see another user's payment), admin-only routes, refund rules, **10 concurrent refund requests → exactly 1 refund**, **50 concurrent admin updates → exactly 1 wins**, dashboard totals, PDF receipt structure and CORS.

## API

| Method | Path | Access |
|---|---|---|
| POST | `/auth/register`, `/auth/login` | public |
| GET | `/auth/me` | user |
| POST, GET | `/transactions/` | user |
| GET | `/transactions/{transaction_id}` | user |
| POST | `/refunds/` | user |
| GET | `/refunds/{transaction_id}` | user |
| POST, GET | `/tickets/` | user |
| GET | `/tickets/{ticket_id}` | user |
| GET | `/dashboard/` | user |
| GET | `/receipt/{transaction_id}` (PDF) | user |
| POST | `/assistant/chat` | user |
| GET | `/admin/dashboard`, `/admin/users`, `/admin/transactions`, `/admin/refunds`, `/admin/tickets` | admin |
| GET | `/refunds/admin/all`, `/tickets/admin/all` | admin |
| PATCH | `/admin/transactions/{id}/status`, `/refunds/admin/{id}/status`, `/tickets/admin/{id}/status` | admin |
| GET | `/`, `/health` | public |

## Deploy (Railway or any Docker host)

A `Dockerfile` is included. Set the root directory to `backend/`, add the environment variables above, and deploy. The final image is a small static binary.
