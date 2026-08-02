"""
FastAPI application entrypoint.
Creates the app, registers routers, sets up CORS, and creates DB
tables on startup.
"""
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.database.db import Base, engine
from app.routes import (
    auth_routes,
    transaction_routes,
    refund_routes,
    dashboard_routes,
    assistant_routes,
    support_ticket_routes,
    admin_routes,
    receipt_routes,
)

# Creates all tables defined in models/ if they don't already exist
Base.metadata.create_all(bind=engine)

app = FastAPI(title="AI Payment Support System")

# CORS: allow the local dev frontend and the deployed frontend.
# Update allow_origins with your actual Vercel URL after deployment.
app.add_middleware(
    CORSMiddleware,
    allow_origins=[
        "http://localhost:5173",
        "https://ai-payment-support.vercel.app",
    ],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(auth_routes.router)
app.include_router(transaction_routes.router)
app.include_router(refund_routes.router)
app.include_router(dashboard_routes.router)
app.include_router(assistant_routes.router)
app.include_router(support_ticket_routes.router)
app.include_router(admin_routes.router)
app.include_router(receipt_routes.router)


@app.get("/")
def root():
    return {"message": "AI Payment Support System API is running"}
