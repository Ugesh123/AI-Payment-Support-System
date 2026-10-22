# 💳 AI Payment Support & Transaction Portal

A full-stack fintech-inspired payment support platform where users can securely register, simulate payments, manage transactions, request refunds, raise support tickets, and monitor payment activity through an interactive dashboard.

Built with a **Golang** backend, a **React** frontend and **MySQL**, using JWT authentication and REST APIs.

**At a glance:** Go backend · 24 REST endpoints · 26 automated tests (run with the race detector) · React + Tailwind frontend · MySQL

---

## 🚀 Features

- 🔐 Secure JWT Authentication
- 👤 User Registration & Login
- 💳 Simulated Payment Transactions
- 📊 Interactive Dashboard
- 📜 Transaction History
- 💰 Refund Request & Tracking
- 🎫 Support Ticket Management
- 👨‍💼 Role-Based Admin Panel
- 📄 PDF Receipt Generation
- 📱 Fully Responsive UI
- 🤖 AI Assistant Module *(Currently Under Maintenance)*

---

## 🛠 Tech Stack

### Backend
- Go 1.22+ (standard library `net/http` router)
- MySQL with `database/sql` and parameterised SQL (`go-sql-driver/mysql`)
- JWT authentication (`golang-jwt/jwt`)
- PBKDF2-SHA256 password hashing
- PDF receipts built with the Go standard library
- `go test` with the race detector, GitHub Actions CI, Docker

### Frontend
- React.js
- Vite
- React Router
- Axios
- Tailwind CSS

### Database
- MySQL

---

## 🧠 Engineering Highlights

- **No floating-point money:** amounts are stored as whole paise, so `0.10 + 0.20` is always exactly `0.30`.
- **No duplicate refunds:** a unique constraint means 10 simultaneous refund requests for one payment create exactly 1 refund (tested).
- **Safe admin updates:** status changes use compare-and-set (`UPDATE ... WHERE status = ?`), so if two admins act at once, only one wins and the other gets `409 Conflict`.
- **State machines:** Refund `REQUESTED → PROCESSING → COMPLETED`; Ticket `OPEN → IN_PROGRESS → CLOSED`; only `PENDING` transactions can change. Every other move is rejected.
- **Ownership checks:** a user can never see another user's payments, refunds or tickets.
- **Production basics:** graceful shutdown, server timeouts, panic recovery and JSON request logs.

---

## 📂 Project Structure

```text
AI-Payment-Support-System/
│
├── backend/                 # Go REST API
│   ├── cmd/server/          # entry point
│   └── internal/            # api, service, domain, store, auth, money, receipt, config
├── frontend/                # React app
├── docs/                    # database schema, sample data, Postman collection
└── README.md
```

Request flow: `api` (HTTP handlers) → `service` (business rules) → `store` (MySQL).

---

## ⚙️ Installation

### Backend

Requires Go 1.22 or newer.

```bash
cd backend
go mod tidy
cp .env.example .env        # then set JWT_SECRET_KEY and DATABASE_URL
go run ./cmd/server
```

Backend runs on

```
http://localhost:8080
```

Run the tests

```bash
go test -race ./...
```

Tip: leave `DATABASE_URL` empty to run without MySQL (data is kept in memory and lost on restart).

---

### Frontend

```bash
cd frontend

npm install

npm run dev
```

Frontend runs on

```
http://localhost:5173
```

---

## 🔐 Environment Variables

### Backend (.env)

```env
JWT_SECRET_KEY=your_long_secret_key
DATABASE_URL=mysql://user:password@localhost:3306/payment_support_db
PORT=8080
ACCESS_TOKEN_EXPIRE_MINUTES=60
CORS_ORIGINS=http://localhost:5173
ADMIN_EMAIL=admin@example.com      # optional: creates an admin on startup
ADMIN_PASSWORD=change-me           # optional
```

### Frontend (.env)

```env
VITE_API_BASE_URL=http://localhost:8080
```

---

## 📌 Main Features

### Authentication

- User Registration
- Secure Login
- JWT Authentication
- Protected Routes

### Payments

- Simulated Payments
- Success / Pending / Failed Status
- Transaction History

### Refunds

- Request Refund
- Refund Status Tracking

### Support

- Raise Support Tickets
- Ticket Status Tracking

### Admin Panel

- Manage Users
- Manage Transactions
- Manage Refunds
- Manage Support Tickets

### AI Assistant

The AI Assistant module is currently under maintenance and will be upgraded in a future release.

---

## 🌐 API Endpoints

| Method | Endpoint | Description |
|---------|----------|-------------|
| POST | /auth/register | Register User |
| POST | /auth/login | Login |
| GET | /auth/me | Current User |
| GET | /dashboard | Dashboard |
| GET | /transactions | List Transactions |
| POST | /transactions | Create Transaction |
| GET | /transactions/{id} | Transaction Details |
| POST | /refunds | Request Refund |
| GET | /refunds/{transaction_id} | Refund Status |
| GET | /tickets | List Tickets |
| POST | /tickets | Create Ticket |
| GET | /receipt/{id} | Download Receipt (PDF) |
| POST | /assistant/chat | AI Assistant |
| GET | /admin/dashboard | Admin Summary |
| GET | /admin/users, /admin/transactions, /admin/refunds, /admin/tickets | Admin Lists |
| PATCH | /admin/transactions/{id}/status | Update Transaction Status |
| PATCH | /refunds/admin/{id}/status | Update Refund Status |
| PATCH | /tickets/admin/{id}/status | Update Ticket Status |

Full list in [`backend/README.md`](backend/README.md).

---

## 🚀 Future Improvements

- AI Assistant Integration
- Razorpay Payment Gateway
- Email Notifications
- SMS Notifications
- Cloud Deployment
- Analytics Dashboard

---

## ⭐ GitHub Repository

https://github.com/Ugesh123/AI-Payment-Support-System

---

## 👨‍💻 Author

**Ugesh**

GitHub:
https://github.com/Ugesh123

LinkedIn:
https://www.linkedin.com/in/yada-ugesh-ba602a345

---

## 📄 License

MIT License
