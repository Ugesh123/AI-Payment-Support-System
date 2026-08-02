"""
Pydantic schemas for support ticket requests/responses.
"""
from pydantic import BaseModel
from datetime import datetime
from typing import Optional
from app.models.support_ticket import TicketStatus


class TicketCreate(BaseModel):
    transaction_id: str   # public TXN-XXXXXXXX id, same pattern as RefundCreate
    subject: str
    description: str


class TicketResponse(BaseModel):
    id: int
    ticket_id: str
    transaction_id: int
    subject: str
    description: str
    status: str
    created_at: datetime
    closed_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class TicketStatusUpdate(BaseModel):
    """Used by the admin to move a ticket forward or close it."""
    status: TicketStatus
