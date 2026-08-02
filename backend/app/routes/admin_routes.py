"""
Single admin panel router -- brings together user, transaction, refund,
and ticket visibility for the admin. Every route here requires the
admin role via get_current_admin().
"""
from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session
from app.database.db import get_db
from app.schemas.user_schema import UserResponse
from app.schemas.transaction_schema import TransactionResponse, TransactionStatusUpdate
from app.schemas.refund_schema import RefundResponse
from app.schemas.ticket_schema import TicketResponse
from app.schemas.admin_schema import AdminDashboardSummary
from app.services import admin_service, transaction_service, refund_service, support_ticket_service
from app.services.dependencies import get_current_admin
from app.models.user import User

router = APIRouter(prefix="/admin", tags=["Admin Panel"])


@router.get("/dashboard", response_model=AdminDashboardSummary)
def get_admin_dashboard(
    db: Session = Depends(get_db),
    admin: User = Depends(get_current_admin)
):
    """Admin: one-call summary for the admin dashboard landing page."""
    return admin_service.get_admin_dashboard_summary(db)


@router.get("/users", response_model=list[UserResponse])
def view_all_users(
    db: Session = Depends(get_db),
    admin: User = Depends(get_current_admin)
):
    """Admin: view all registered users."""
    return admin_service.get_all_users(db)


@router.get("/transactions", response_model=list[TransactionResponse])
def view_all_transactions(
    db: Session = Depends(get_db),
    admin: User = Depends(get_current_admin)
):
    """Admin: view all transactions across all users."""
    return transaction_service.get_all_transactions(db)


@router.patch("/transactions/{transaction_id}/status", response_model=TransactionResponse)
def update_transaction_status(
    transaction_id: int,
    status_update: TransactionStatusUpdate,
    db: Session = Depends(get_db),
    admin: User = Depends(get_current_admin)
):
    """Admin: update a PENDING transaction's status to SUCCESS or FAILED."""
    return transaction_service.update_transaction_status(db, transaction_id, status_update.status)


@router.get("/refunds", response_model=list[RefundResponse])
def view_all_refunds(
    db: Session = Depends(get_db),
    admin: User = Depends(get_current_admin)
):
    """Admin: view all refund requests. Reuses refund_service."""
    return refund_service.get_all_refunds(db)


@router.get("/tickets", response_model=list[TicketResponse])
def view_all_tickets(
    db: Session = Depends(get_db),
    admin: User = Depends(get_current_admin)
):
    """Admin: view all support tickets. Reuses support_ticket_service."""
    return support_ticket_service.get_all_tickets(db)
