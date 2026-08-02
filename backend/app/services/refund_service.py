"""
Business logic for requesting, viewing, and updating refunds.
"""
from datetime import datetime, timezone
from sqlalchemy.orm import Session
from fastapi import HTTPException, status as http_status
from app.models.transaction import Transaction, TransactionStatus
from app.models.refund import Refund, RefundStatus
from app.models.user import User
from app.schemas.refund_schema import RefundCreate

# Fixed estimate shown to users -- no real bank/gateway timing to calculate.
EXPECTED_REFUND_DAYS = 3

# Defines the ONLY allowed forward moves. Anything not listed here is invalid.
VALID_STATUS_TRANSITIONS = {
    RefundStatus.REQUESTED: [RefundStatus.PROCESSING],
    RefundStatus.PROCESSING: [RefundStatus.COMPLETED],
    RefundStatus.COMPLETED: [],   # COMPLETED is final -- no further moves allowed
}


def request_refund(db: Session, user: User, refund_data: RefundCreate) -> Refund:
    """
    Creates a refund request for the full transaction amount.
    Only allowed for the user's own SUCCESSFUL transactions,
    and only once per transaction.
    """
    txn = db.query(Transaction).filter(
        Transaction.transaction_id == refund_data.transaction_id,
        Transaction.user_id == user.id
    ).first()

    if not txn:
        raise HTTPException(status_code=http_status.HTTP_404_NOT_FOUND, detail="Transaction not found")

    if txn.status != TransactionStatus.SUCCESS:
        raise HTTPException(
            status_code=http_status.HTTP_400_BAD_REQUEST,
            detail="Refund can only be requested for successful transactions"
        )

    if txn.refund:
        raise HTTPException(
            status_code=http_status.HTTP_400_BAD_REQUEST,
            detail="A refund has already been requested for this transaction"
        )

    new_refund = Refund(
        transaction_id=txn.id,
        amount=txn.amount,
        reason=refund_data.refund_reason,
        status=RefundStatus.REQUESTED,
    )

    db.add(new_refund)
    db.commit()
    db.refresh(new_refund)
    return new_refund


def get_refund_status(db: Session, user: User, transaction_id: str) -> Refund:
    """Fetches the refund status for a transaction, only if it belongs to the user."""
    txn = db.query(Transaction).filter(
        Transaction.transaction_id == transaction_id,
        Transaction.user_id == user.id
    ).first()

    if not txn or not txn.refund:
        raise HTTPException(status_code=http_status.HTTP_404_NOT_FOUND, detail="Refund not found")

    return txn.refund


# ---------- Admin-only functions ----------

def get_all_refunds(db: Session) -> list[Refund]:
    """Admin: view every refund request in the system."""
    return db.query(Refund).order_by(Refund.requested_at.desc()).all()


def update_refund_status(db: Session, refund_id: int, new_status: RefundStatus) -> Refund:
    """
    Admin: move a refund forward one step at a time.
    Only REQUESTED -> PROCESSING -> COMPLETED is allowed.
    Rejects both backward/invalid jumps AND setting the same status twice.
    """
    refund = db.query(Refund).filter(Refund.id == refund_id).first()
    if not refund:
        raise HTTPException(status_code=http_status.HTTP_404_NOT_FOUND, detail="Refund not found")

    if refund.status == new_status:
        raise HTTPException(
            status_code=http_status.HTTP_400_BAD_REQUEST,
            detail=f"Refund is already in {new_status.value} status."
        )

    allowed_next_statuses = VALID_STATUS_TRANSITIONS[refund.status]
    if new_status not in allowed_next_statuses:
        raise HTTPException(
            status_code=http_status.HTTP_400_BAD_REQUEST,
            detail=f"Invalid refund status transition: {refund.status.value} -> {new_status.value}"
        )

    refund.status = new_status
    if new_status == RefundStatus.COMPLETED:
        refund.completed_at = datetime.now(timezone.utc)

    db.commit()
    db.refresh(refund)
    return refund
