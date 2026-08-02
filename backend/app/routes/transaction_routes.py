"""
API endpoints for transactions. All routes require a logged-in user (JWT).
"""
from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session
from app.database.db import get_db
from app.schemas.transaction_schema import TransactionCreate, TransactionResponse
from app.services import transaction_service
from app.services.dependencies import get_current_user
from app.models.user import User

router = APIRouter(prefix="/transactions", tags=["Transactions"])

@router.post("/", response_model=TransactionResponse)
def create_transaction(
    txn_data: TransactionCreate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """Simulates a payment ('Pay Now' button) and saves the result."""
    return transaction_service.create_transaction(db, current_user, txn_data)


@router.get("/", response_model=list[TransactionResponse])
def list_transactions(
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """Returns the logged-in user's transaction history."""
    return transaction_service.get_user_transactions(db, current_user)


@router.get("/{transaction_id}", response_model=TransactionResponse)
def get_transaction(
    transaction_id: str,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """Search for a specific transaction by its transaction_id."""
    return transaction_service.get_transaction_by_txn_id(db, current_user, transaction_id)
