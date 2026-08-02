# 💳 AI Payment Support & Transaction Portal

A full-stack fintech-inspired payment support platform where users can securely register, simulate payments, manage transactions, request refunds, raise support tickets, and monitor payment activity through an interactive dashboard.

This project demonstrates full-stack software engineering using **FastAPI, React, MySQL, JWT Authentication, SQLAlchemy, and REST APIs**.

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
- Python
- FastAPI
- SQLAlchemy
- MySQL
- JWT Authentication
- ReportLab

### Frontend
- React.js
- Vite
- React Router
- Axios
- Tailwind CSS

### Database
- MySQL

---

## 📂 Project Structure

```text
AI-Payment-Support-System/
│
├── backend/
├── frontend/
├── screenshots/
├── docs/
└── README.md
```

---

## ⚙️ Installation

### Backend

```bash
cd backend

python -m venv venv

# Windows
venv\Scripts\activate

pip install -r requirements.txt

uvicorn app.main:app --reload
```

Backend runs on

```
http://localhost:8000
```

Swagger API

```
http://localhost:8000/docs
```

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
DATABASE_URL=your_database_url
JWT_SECRET_KEY=your_secret_key
JWT_ALGORITHM=HS256
ACCESS_TOKEN_EXPIRE_MINUTES=60
```

### Frontend (.env)

```env
VITE_API_BASE_URL=http://127.0.0.1:8000
```

---

## 📸 Screenshots

Add screenshots inside the **screenshots** folder.

Recommended screenshots:

- Login
- Register
- Dashboard
- Transactions
- Refund Request
- Support Tickets
- AI Assistant
- Admin Dashboard
- Swagger API

Example:

```md
## Dashboard

![Dashboard](screenshots/dashboard.png)
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
| POST | /refunds | Request Refund |
| GET | /tickets | List Tickets |
| POST | /tickets | Create Ticket |
| GET | /receipt/{id} | Download Receipt |

Interactive API Documentation:

```
http://localhost:8000/docs
```

---

## 🚀 Future Improvements

- AI Assistant Integration
- Razorpay Payment Gateway
- Email Notifications
- SMS Notifications
- Cloud Deployment
- Analytics Dashboard
- Unit Testing

---

## ⭐ GitHub Repository

https://github.com/Ugesh123/AI-Payment-Support-System

---

## 👨‍💻 Author

**Ugesh**

GitHub:
https://github.com/Ugesh123

LinkedIn:
(Add your LinkedIn profile)

---

## 📄 License

MIT License
