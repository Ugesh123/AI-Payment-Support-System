"""
Business logic for creating and retrieving transactions.
Includes the simulated payment gateway logic and the admin-only
status correction logic (state machine).
"""
import random
import uuid
from sqlalchemy.orm import Session
from fastapi import HTTPException, status as http_status
from app.models.transaction import Transaction, TransactionStatus
from app.models.user import User
from app.schemas.transaction_schema import TransactionCreate


def generate_transaction_id() -> str:
    """
    Creates a unique transaction ID like 'TXN-9F2A4C8B'.
    Uses the first 8 characters of a UUID4 hex string (uppercased).
    """
    unique_part = uuid.uuid4().hex[:8].upper()
    return f"TXN-{unique_part}"


def simulate_payment_result() -> tuple[str, str | None]:
    """
    Simulates what a real payment gateway would return.
    70% Success, 20% Failed, 10% Pending -- weighted random choice.
    """
    outcome = random.choices(
        population=[TransactionStatus.SUCCESS, TransactionStatus.FAILED, TransactionStatus.PENDING],
        weights=[70, 20, 10],
        k=1
    )[0]

    failure_reasons = [
        "Insufficient Balance",
        "Bank Server Timeout",
        "Card Declined",
        "Incorrect OTP"
    ]

    reason = random.choice(failure_reasons) if outcome == TransactionStatus.FAILED else None
    return outcome, reason


def create_transaction(db: Session, user: User, txn_data: TransactionCreate) -> Transaction:
    """
    Creates a new transaction. Amount and payment_method are already
    validated by the Pydantic schema before this function runs.
    """
    result_status, failure_reason = simulate_payment_result()

    new_txn = Transaction(
        transaction_id=generate_transaction_id(),
        user_id=user.id,
        amount=txn_data.amount,
        status=result_status,
        failure_reason=failure_reason,
        payment_method=txn_data.payment_method,
        receiver_name=txn_data.receiver_name,
    )

    db.add(new_txn)
    db.commit()
    db.refresh(new_txn)
    return new_txn


def get_user_transactions(db: Session, user: User) -> list[Transaction]:
    """Returns all transactions belonging to the logged-in user, newest first."""
    return db.query(Transaction).filter(Transaction.user_id == user.id).order_by(Transaction.created_at.desc()).all()


def get_transaction_by_txn_id(db: Session, user: User, transaction_id: str) -> Transaction:
    """
    Fetches a single transaction by its public transaction_id.
    Ensures a user can only view their own transaction.
    """
    txn = db.query(Transaction).filter(
        Transaction.transaction_id == transaction_id,
        Transaction.user_id == user.id
    ).first()

    if not txn:
        raise HTTPException(status_code=http_status.HTTP_404_NOT_FOUND, detail="Transaction not found")

    return txn


# ---------- Admin-only functions ----------

# Admin can correct a transaction's outcome, e.g. a PENDING payment
# that the (simulated) bank later confirmed as SUCCESS or FAILED.
# SUCCESS and FAILED are treated as final -- no further changes allowed.
VALID_TRANSACTION_STATUS_TRANSITIONS = {
    TransactionStatus.PENDING: [TransactionStatus.SUCCESS, TransactionStatus.FAILED],
    TransactionStatus.SUCCESS: [],
    TransactionStatus.FAILED: [],
}


def update_transaction_status(db: Session, transaction_id: int, new_status: TransactionStatus) -> Transaction:
    """
    Admin: update a transaction's status, only while it's still PENDING.
    Once a transaction is SUCCESS or FAILED, it cannot be changed.
    """
    txn = db.query(Transaction).filter(Transaction.id == transaction_id).first()
    if not txn:
        raise HTTPException(status_code=http_status.HTTP_404_NOT_FOUND, detail="Transaction not found")

    if txn.status == new_status:
        raise HTTPException(
            status_code=http_status.HTTP_400_BAD_REQUEST,
            detail=f"Transaction is already in {new_status.value} status."
        )

    allowed_next_statuses = VALID_TRANSACTION_STATUS_TRANSITIONS[txn.status]
    if new_status not in allowed_next_statuses:
        raise HTTPException(
            status_code=http_status.HTTP_400_BAD_REQUEST,
            detail=f"Invalid transaction status transition: {txn.status.value} -> {new_status.value}"
        )

    txn.status = new_status
    db.commit()
    db.refresh(txn)
    return txn


def get_all_transactions(db: Session) -> list[Transaction]:
    """Admin: view every transaction in the system, across all users."""
    return db.query(Transaction).order_by(Transaction.created_at.desc()).all()
