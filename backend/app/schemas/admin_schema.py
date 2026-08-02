"""
Pydantic schema for the admin dashboard summary.
Kept separate from dashboard_schema.py since this summarizes the
whole system, not a single user's data.
"""
from pydantic import BaseModel

class AdminDashboardSummary(BaseModel):
    total_users: int
    total_transactions: int
    pending_transactions: int
    active_refunds: int
    open_tickets: int
