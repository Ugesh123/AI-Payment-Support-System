"""
Admin-only service functions that don't naturally belong to another module.
Viewing refunds and tickets already exists in their own services -- this
file adds what's missing: listing users and the system-wide dashboard summary.
"""
from sqlalchemy.orm import Session
from app.models.user import User
from app.models.transaction import Transaction, TransactionStatus
from app.models.refund import Refund, RefundStatus
from app.models.support_ticket import SupportTicket, TicketStatus
from app.schemas.admin_schema import AdminDashboardSummary


def get_all_users(db: Session) -> list[User]:
    """Admin: view every registered user."""
    return db.query(User).order_by(User.created_at.desc()).all()


def get_admin_dashboard_summary(db: Session) -> AdminDashboardSummary:
    """
    Single-call summary for the admin dashboard, so the React admin
    screen doesn't need to fire five separate requests on load.
    """
    total_users = db.query(User).count()
    total_transactions = db.query(Transaction).count()
    pending_transactions = db.query(Transaction).filter(
        Transaction.status == TransactionStatus.PENDING
    ).count()
    active_refunds = db.query(Refund).filter(
        Refund.status.in_([RefundStatus.REQUESTED, RefundStatus.PROCESSING])
    ).count()
    open_tickets = db.query(SupportTicket).filter(
        SupportTicket.status.in_([TicketStatus.OPEN, TicketStatus.IN_PROGRESS])
    ).count()

    return AdminDashboardSummary(
        total_users=total_users,
        total_transactions=total_transactions,
        pending_transactions=pending_transactions,
        active_refunds=active_refunds,
        open_tickets=open_tickets,
    )
