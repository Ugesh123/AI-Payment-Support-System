"""
SQLAlchemy model for the 'support_tickets' table.
A ticket is always tied to one of the user's own transactions
(e.g. "I have an issue with this specific payment").
"""
from sqlalchemy import Column, Integer, String, Text, DateTime, ForeignKey, Enum
from sqlalchemy.sql import func
from sqlalchemy.orm import relationship
from app.database.db import Base
import enum


class TicketStatus(str, enum.Enum):
    OPEN = "OPEN"
    IN_PROGRESS = "IN_PROGRESS"
    CLOSED = "CLOSED"


class SupportTicket(Base):
    __tablename__ = "support_tickets"

    id = Column(Integer, primary_key=True, index=True)
    ticket_id = Column(String(50), unique=True, index=True, nullable=False)   # e.g. TKT-9F2A4C8B

    user_id = Column(Integer, ForeignKey("users.id"), nullable=False)
    transaction_id = Column(Integer, ForeignKey("transactions.id"), nullable=False)

    subject = Column(String(150), nullable=False)
    description = Column(Text, nullable=False)
    status = Column(Enum(TicketStatus), nullable=False, default=TicketStatus.OPEN)

    created_at = Column(DateTime(timezone=True), server_default=func.now())
    closed_at = Column(DateTime(timezone=True), nullable=True)

    user = relationship("User")
    transaction = relationship("Transaction")
