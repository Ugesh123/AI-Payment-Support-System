"""
Pydantic schemas for the dashboard response.
"""
from pydantic import BaseModel
from decimal import Decimal
from typing import List

class DashboardSummary(BaseModel):
    total_transactions: int
    successful_payments: int
    failed_payments: int
    pending_payments: int
    total_amount_paid: Decimal
    active_refunds: int
    completed_refunds: int

class RecentTransaction(BaseModel):
    transaction_id: str
    receiver: str
    amount: Decimal
    status: str

class DashboardResponse(BaseModel):
    summary: DashboardSummary
    recent_transactions: List[RecentTransaction]
