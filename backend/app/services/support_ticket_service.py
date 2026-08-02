"""
Business logic for creating, viewing, and managing support tickets.
Follows the same state-machine pattern used in refund_service.py.
"""
import uuid
from datetime import datetime, timezone
from sqlalchemy.orm import Session
from fastapi import HTTPException, status as http_status
from app.models.support_ticket import SupportTicket, TicketStatus
from app.models.transaction import Transaction
from app.models.user import User
from app.schemas.ticket_schema import TicketCreate

# Defines the ONLY allowed forward moves for a ticket's status.
VALID_STATUS_TRANSITIONS = {
    TicketStatus.OPEN: [TicketStatus.IN_PROGRESS, TicketStatus.CLOSED],
    TicketStatus.IN_PROGRESS: [TicketStatus.CLOSED],
    TicketStatus.CLOSED: [],
}


def generate_ticket_id() -> str:
    """Creates a unique ticket ID like 'TKT-9F2A4C8B', same style as transaction IDs."""
    unique_part = uuid.uuid4().hex[:8].upper()
    return f"TKT-{unique_part}"


def create_ticket(db: Session, user: User, ticket_data: TicketCreate) -> SupportTicket:
    """
    Creates a support ticket for one of the user's own transactions.
    """
    txn = db.query(Transaction).filter(
        Transaction.transaction_id == ticket_data.transaction_id,
        Transaction.user_id == user.id
    ).first()

    if not txn:
        raise HTTPException(status_code=http_status.HTTP_404_NOT_FOUND, detail="Transaction not found")

    new_ticket = SupportTicket(
        ticket_id=generate_ticket_id(),
        user_id=user.id,
        transaction_id=txn.id,
        subject=ticket_data.subject,
        description=ticket_data.description,
        status=TicketStatus.OPEN,
    )

    db.add(new_ticket)
    db.commit()
    db.refresh(new_ticket)
    return new_ticket


def get_user_tickets(db: Session, user: User) -> list[SupportTicket]:
    """Returns all tickets raised by the logged-in user, newest first."""
    return (
        db.query(SupportTicket)
        .filter(SupportTicket.user_id == user.id)
        .order_by(SupportTicket.created_at.desc())
        .all()
    )


def get_ticket_by_id(db: Session, user: User, ticket_id: str) -> SupportTicket:
    """Fetches a single ticket, only if it belongs to the logged-in user."""
    ticket = db.query(SupportTicket).filter(
        SupportTicket.ticket_id == ticket_id,
        SupportTicket.user_id == user.id
    ).first()

    if not ticket:
        raise HTTPException(status_code=http_status.HTTP_404_NOT_FOUND, detail="Ticket not found")

    return ticket


# ---------- Admin-only functions ----------

def get_all_tickets(db: Session) -> list[SupportTicket]:
    """Admin: view every support ticket in the system."""
    return db.query(SupportTicket).order_by(SupportTicket.created_at.desc()).all()


def update_ticket_status(db: Session, ticket_id: int, new_status: TicketStatus) -> SupportTicket:
    """
    Admin: move a ticket forward (OPEN -> IN_PROGRESS -> CLOSED),
    or close it directly from OPEN.
    """
    ticket = db.query(SupportTicket).filter(SupportTicket.id == ticket_id).first()
    if not ticket:
        raise HTTPException(status_code=http_status.HTTP_404_NOT_FOUND, detail="Ticket not found")

    if ticket.status == new_status:
        raise HTTPException(
            status_code=http_status.HTTP_400_BAD_REQUEST,
            detail=f"Ticket is already in {new_status.value} status."
        )

    allowed_next_statuses = VALID_STATUS_TRANSITIONS[ticket.status]
    if new_status not in allowed_next_statuses:
        raise HTTPException(
            status_code=http_status.HTTP_400_BAD_REQUEST,
            detail=f"Invalid ticket status transition: {ticket.status.value} -> {new_status.value}"
        )

    ticket.status = new_status
    if new_status == TicketStatus.CLOSED:
        ticket.closed_at = datetime.now(timezone.utc)

    db.commit()
    db.refresh(ticket)
    return ticket
