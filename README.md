# AI Payment Support & Transaction Portal

A full-stack fintech-style support platform where users can simulate payments,
track transactions, request refunds, raise support tickets, and get AI-powered
explanations of their payment status — plus a role-based admin panel to manage
the whole system.

Built as a portfolio project demonstrating backend engineering and AI
integration skills relevant to fintech roles (e.g. Cashfree, Razorpay).

## Features

- **JWT Authentication** — Register/Login, token verified via `/auth/me` on every app load
- **Simulated Payments** — weighted random SUCCESS/FAILED/PENDING outcomes (no real gateway)
- **Transaction History** — search, status filtering, pagination
- **Refund Requests** — REQUESTED → PROCESSING → COMPLETED state machine
- **AI Payment Assistant** — Google Gemini, grounded in real transaction data from MySQL
- **Support Tickets** — OPEN → IN_PROGRESS → CLOSED workflow
- **Role-Based Admin Panel** — manage users, transactions, refunds, and tickets
- **PDF Receipt Generation** — for successful payments only
- **Fully Responsive UI** — mobile, tablet, and desktop

## Tech Stack

**Backend:** Python, FastAPI, SQLAlchemy, MySQL, JWT, ReportLab, Google Gemini API
**Frontend:** React, Vite, React Router, Axios, Tailwind CSS
**Database:** MySQL
**Deployment:** Backend on Railway, Frontend on Vercel

## Project Structure

```
AI-Payment-Support-System/
├── backend/
│   ├── app/
│   │   ├── main.py
│   │   ├── config.py
│   │   ├── database/
│   │   ├── models/
│   │   ├── schemas/
│   │   ├── routes/
│   │   ├── services/
│   │   └── utils/
│   ├── requirements.txt
│   ├── .env.example
│   ├── .gitignore
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── api/
│   │   ├── components/
│   │   ├── context/
│   │   ├── hooks/
│   │   ├── pages/
│   │   ├── services/
│   │   └── utils/
│   ├── package.json
│   ├── .env.example
│   └── .gitignore
├── docs/
│   ├── database_schema.sql
│   ├── sample_data.sql
│   ├── postman_collection.json
│   ├── screenshots/
│   └── demo/
└── README.md
```

## Installation

### Backend

```bash
cd backend
python -m venv venv
source venv/bin/activate   # venv\Scripts\activate on Windows
pip install -r requirements.txt
cp .env.example .env       # fill in your DATABASE_URL, JWT_SECRET_KEY, GEMINI_API_KEY
uvicorn app.main:app --reload
```

Visit `http://localhost:8000/docs` for the interactive Swagger UI.

### Frontend

```bash
cd frontend
npm install
cp .env.example .env       # set VITE_API_BASE_URL=http://localhost:8000
npm run dev
```

Visit `http://localhost:5173`.

### Database

Run the schema and (optionally) sample data against your MySQL instance:

```bash
mysql -u <user> -p < docs/database_schema.sql
mysql -u <user> -p < docs/sample_data.sql
```

To create your first admin account: register normally through the app, then
promote the user directly in MySQL:

```sql
UPDATE users SET role = 'admin' WHERE email = 'youremail@example.com';
```

## Environment Variables

**Backend (`backend/.env`)**
```env
DATABASE_URL=mysql+pymysql://user:password@host:port/dbname
JWT_SECRET_KEY=your-long-random-secret
JWT_ALGORITHM=HS256
ACCESS_TOKEN_EXPIRE_MINUTES=60
GEMINI_API_KEY=your-gemini-key
```

**Frontend (`frontend/.env`)**
```env
VITE_API_BASE_URL=http://localhost:8000
```

## API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| POST | /auth/register | Register a new user |
| POST | /auth/login | Login, returns JWT |
| GET | /auth/me | Get current logged-in user |
| POST | /transactions/ | Simulate a payment |
| GET | /transactions/ | List user's transactions |
| GET | /transactions/{id} | Get one transaction |
| POST | /refunds/ | Request a refund |
| GET | /refunds/{transaction_id} | Get refund status |
| GET | /dashboard/ | User dashboard summary |
| POST | /assistant/chat | Ask the AI assistant |
| POST | /tickets/ | Create a support ticket |
| GET | /tickets/ | List user's tickets |
| GET | /receipt/{transaction_id} | Download PDF receipt (SUCCESS only) |
| GET | /admin/dashboard | Admin summary (admin only) |
| GET | /admin/users | List all users (admin only) |
| GET | /admin/transactions | List all transactions (admin only) |
| PATCH | /admin/transactions/{id}/status | Update transaction status (admin only) |
| GET | /admin/refunds | List all refunds (admin only) |
| PATCH | /refunds/admin/{id}/status | Update refund status (admin only) |
| GET | /admin/tickets | List all tickets (admin only) |
| PATCH | /tickets/admin/{id}/status | Update ticket status (admin only) |

Full interactive documentation is available at `/docs` (Swagger) or `/redoc`.
A ready-to-import Postman collection is included at `docs/postman_collection.json`.

## Screenshots

Add screenshots to `docs/screenshots/` — recommended: Login, Dashboard,
Transactions, AI Assistant chat, and Admin Panel.

## Deployment Guide

### Backend → Railway
1. Push the repo to GitHub, create a new Railway project from it (point at `backend/` if monorepo).
2. Start command: `uvicorn app.main:app --host 0.0.0.0 --port $PORT`
3. Add the environment variables listed above under Railway's **Variables** tab.
4. Note your deployed URL, e.g. `https://ai-payment-backend.up.railway.app`.

### Database → Railway MySQL
Provision a MySQL instance on Railway and use its connection string as `DATABASE_URL`.
The same instance can be used for both development and production.

### Frontend → Vercel
1. Import the repo into Vercel, set **Root Directory** to `frontend`.
2. Framework preset: Vite.
3. Add `VITE_API_BASE_URL` pointing to your deployed backend URL.
4. Deploy.

### Final CORS step
After deploying, update `allow_origins` in `backend/app/main.py` to include your
live Vercel URL, then redeploy the backend.

## Future Improvements

- Server-side pagination for transactions (currently client-side, fine at current scale)
- Email notifications on refund/ticket status changes
- Chat history persistence for the AI Assistant (currently session-only)
- Automated tests (pytest for backend, React Testing Library for frontend)
- Rate limiting on authentication endpoints

## License

MIT
