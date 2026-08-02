"""
Business logic for building the dashboard summary and recent transactions.
Uses SQL aggregate functions (COUNT, SUM) instead of pulling all rows
into Python and counting manually -- faster and more efficient.
"""
from decimal import Decimal
from sqlalchemy import func
from sqlalchemy.orm import Session
from app.models.transaction import Transaction, TransactionStatus
from app.models.refund import Refund, RefundStatus
from app.models.user import User
from app.schemas.dashboard_schema import DashboardSummary, RecentTransaction, DashboardResponse


def get_dashboard_summary(db: Session, user: User) -> DashboardSummary:
    base_query = db.query(Transaction).filter(Transaction.user_id == user.id)

    total_transactions = base_query.count()
    successful_payments = base_query.filter(Transaction.status == TransactionStatus.SUCCESS).count()
    failed_payments = base_query.filter(Transaction.status == TransactionStatus.FAILED).count()
    pending_payments = base_query.filter(Transaction.status == TransactionStatus.PENDING).count()

    total_amount_paid = (
        db.query(func.sum(Transaction.amount))
        .filter(Transaction.user_id == user.id, Transaction.status == TransactionStatus.SUCCESS)
        .scalar() or Decimal("0.00")
    )

    refund_query = db.query(Refund).join(Transaction).filter(Transaction.user_id == user.id)
    active_refunds = refund_query.filter(
        Refund.status.in_([RefundStatus.REQUESTED, RefundStatus.PROCESSING])
    ).count()
    completed_refunds = refund_query.filter(Refund.status == RefundStatus.COMPLETED).count()

    return DashboardSummary(
        total_transactions=total_transactions,
        successful_payments=successful_payments,
        failed_payments=failed_payments,
        pending_payments=pending_payments,
        total_amount_paid=total_amount_paid,
        active_refunds=active_refunds,
        completed_refunds=completed_refunds,
    )


def get_recent_transactions(db: Session, user: User, limit: int = 5) -> list[RecentTransaction]:
    """Returns the latest N transactions for the user, newest first."""
    transactions = (
        db.query(Transaction)
        .filter(Transaction.user_id == user.id)
        .order_by(Transaction.created_at.desc())
        .limit(limit)
        .all()
    )

    return [
        RecentTransaction(
            transaction_id=t.transaction_id,
            receiver=t.receiver_name,
            amount=t.amount,
            status=t.status.value,
        )
        for t in transactions
    ]


def get_dashboard(db: Session, user: User) -> DashboardResponse:
    """Combines the summary and recent transactions into one response."""
    return DashboardResponse(
        summary=get_dashboard_summary(db, user),
        recent_transactions=get_recent_transactions(db, user),
    )
