"""
Pydantic schemas for transaction requests/responses.
Uses Decimal (not float) to match the database precision,
and validates amount + payment_method before the request even
reaches the service layer.
"""
from pydantic import BaseModel, field_validator
from decimal import Decimal
from datetime import datetime
from typing import Optional
from app.models.transaction import PaymentMethod, TransactionStatus

MIN_AMOUNT = Decimal("1.00")
MAX_AMOUNT = Decimal("100000.00")

class TransactionCreate(BaseModel):
    amount: Decimal
    receiver_name: str
    payment_method: PaymentMethod

    @field_validator("amount")
    @classmethod
    def validate_amount(cls, value: Decimal) -> Decimal:
        if value <= 0:
            raise ValueError("Amount must be greater than zero")
        if not (MIN_AMOUNT <= value <= MAX_AMOUNT):
            raise ValueError(f"Amount must be between {MIN_AMOUNT} and {MAX_AMOUNT}")
        return value

class TransactionResponse(BaseModel):
    id: int
    transaction_id: str
    amount: Decimal
    status: str
    failure_reason: Optional[str] = None
    payment_method: str
    receiver_name: str
    created_at: datetime

    class Config:
        from_attributes = True

class TransactionStatusUpdate(BaseModel):
    """Used by the admin to correct/update a transaction's status."""
    status: TransactionStatus
