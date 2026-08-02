"""
AI Assistant service.

Temporarily disabled.
Returns a maintenance message until AI integration is enabled.
"""

from sqlalchemy.orm import Session
from fastapi import HTTPException, status as http_status

from app.models.transaction import Transaction
from app.models.user import User


def ask_assistant(db: Session, user: User, transaction_id: str, question: str) -> str:
    """
    Temporary AI Assistant response.
    """

    transaction = (
        db.query(Transaction)
        .filter(
            Transaction.transaction_id == transaction_id,
            Transaction.user_id == user.id,
        )
        .first()
    )

    if not transaction:
        raise HTTPException(
            status_code=http_status.HTTP_404_NOT_FOUND,
            detail="Transaction not found",
        )

    return f"""
🚧 AI Assistant is Currently Under Maintenance

Thank you for using the AI Payment Support System.

Our AI Assistant is temporarily unavailable while we upgrade and improve its capabilities.

Transaction ID: {transaction.transaction_id}
Payment Status: {transaction.status.value}

For immediate assistance, you can:
• Check your transaction details.
• Request a refund (if eligible).
• Raise a support ticket.
• Contact customer support.

We appreciate your patience and will bring the AI Assistant back soon.
"""