"""
API endpoints for refunds. User routes require login;
admin routes require the admin role.
"""
from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session
from app.database.db import get_db
from app.schemas.refund_schema import RefundCreate, RefundResponse, RefundStatusUpdate
from app.services import refund_service
from app.services.dependencies import get_current_user, get_current_admin
from app.models.user import User

router = APIRouter(prefix="/refunds", tags=["Refunds"])

@router.post("/", response_model=RefundResponse)
def request_refund(
    refund_data: RefundCreate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """Request a refund for a successful transaction."""
    return refund_service.request_refund(db, current_user, refund_data)


@router.get("/{transaction_id}", response_model=RefundResponse)
def get_refund_status(
    transaction_id: str,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """Check the status of a refund for a given transaction."""
    return refund_service.get_refund_status(db, current_user, transaction_id)


# ---------- Admin-only routes ----------

@router.get("/admin/all", response_model=list[RefundResponse])
def list_all_refunds(
    db: Session = Depends(get_db),
    admin: User = Depends(get_current_admin)
):
    """Admin: view all refund requests across all users."""
    return refund_service.get_all_refunds(db)


@router.patch("/admin/{refund_id}/status", response_model=RefundResponse)
def update_refund_status(
    refund_id: int,
    status_update: RefundStatusUpdate,
    db: Session = Depends(get_db),
    admin: User = Depends(get_current_admin)
):
    """Admin: update a refund's status (REQUESTED -> PROCESSING -> COMPLETED)."""
    return refund_service.update_refund_status(db, refund_id, status_update.status)
