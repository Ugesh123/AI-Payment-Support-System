"""
SQLAlchemy model for the 'transactions' table.
Uses Numeric (not Float) for amount to avoid floating-point precision
errors -- critical for any financial application.
"""
from sqlalchemy import Column, Integer, String, Numeric, DateTime, ForeignKey, Enum
from sqlalchemy.sql import func
from sqlalchemy.orm import relationship
from app.database.db import Base
import enum

class TransactionStatus(str, enum.Enum):
    SUCCESS = "SUCCESS"
    FAILED = "FAILED"
    PENDING = "PENDING"

class PaymentMethod(str, enum.Enum):
    UPI = "UPI"
    CARD = "Card"
    NET_BANKING = "Net Banking"
    WALLET = "Wallet"

class Transaction(Base):
    __tablename__ = "transactions"

    id = Column(Integer, primary_key=True, index=True)
    transaction_id = Column(String(50), unique=True, index=True, nullable=False)
    user_id = Column(Integer, ForeignKey("users.id"), nullable=False)

    # Numeric(10, 2) = up to 10 digits total, 2 after the decimal point.
    # Stored as a fixed-point value in MySQL, no rounding errors like float.
    amount = Column(Numeric(10, 2), nullable=False)

    status = Column(Enum(TransactionStatus), nullable=False)
    failure_reason = Column(String(255), nullable=True)

    payment_method = Column(Enum(PaymentMethod), nullable=False)
    receiver_name = Column(String(100), nullable=False)

    created_at = Column(DateTime(timezone=True), server_default=func.now())

    refund = relationship("Refund", back_populates="transaction", uselist=False)
