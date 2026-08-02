"""
SQLAlchemy model for the 'refunds' table.
Refunds are always for the FULL transaction amount (no partial refunds),
and always start as REQUESTED -- an admin manually moves them forward.
"""
from sqlalchemy import Column, Integer, Numeric, String, DateTime, ForeignKey, Enum
from sqlalchemy.sql import func
from sqlalchemy.orm import relationship
from app.database.db import Base
import enum

class RefundStatus(str, enum.Enum):
    REQUESTED = "REQUESTED"
    PROCESSING = "PROCESSING"
    COMPLETED = "COMPLETED"

class RefundReason(str, enum.Enum):
    DUPLICATE_PAYMENT = "Duplicate Payment"
    WRONG_AMOUNT = "Wrong Amount"
    ORDER_CANCELLED = "Order Cancelled"
    OTHER = "Other"

class Refund(Base):
    __tablename__ = "refunds"

    id = Column(Integer, primary_key=True, index=True)
    transaction_id = Column(Integer, ForeignKey("transactions.id"), nullable=False, unique=True)

    amount = Column(Numeric(10, 2), nullable=False)      # always equals transaction.amount
    reason = Column(Enum(RefundReason), nullable=False)
    status = Column(Enum(RefundStatus), nullable=False, default=RefundStatus.REQUESTED)

    requested_at = Column(DateTime(timezone=True), server_default=func.now())
    completed_at = Column(DateTime(timezone=True), nullable=True)

    transaction = relationship("Transaction", back_populates="refund")
