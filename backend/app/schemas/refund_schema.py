"""
Pydantic schemas for refund requests/responses.
"""
from pydantic import BaseModel
from decimal import Decimal
from datetime import datetime
from typing import Optional
from app.models.refund import RefundReason, RefundStatus

class RefundCreate(BaseModel):
    transaction_id: str          # public TXN-XXXXXXXX id
    refund_reason: RefundReason  # validated against the enum automatically

class RefundResponse(BaseModel):
    id: int
    transaction_id: int
    amount: Decimal
    reason: str
    status: str
    requested_at: datetime
    completed_at: Optional[datetime] = None

    class Config:
        from_attributes = True

class RefundStatusUpdate(BaseModel):
    """Used by the admin to move a refund forward: REQUESTED -> PROCESSING -> COMPLETED"""
    status: RefundStatus
