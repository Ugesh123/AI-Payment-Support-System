"""
API endpoints for support tickets.
User routes require login; admin routes require the admin role.
"""
from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session
from app.database.db import get_db
from app.schemas.ticket_schema import TicketCreate, TicketResponse, TicketStatusUpdate
from app.services import support_ticket_service
from app.services.dependencies import get_current_user, get_current_admin
from app.models.user import User

router = APIRouter(prefix="/tickets", tags=["Support Tickets"])


@router.post("/", response_model=TicketResponse)
def create_ticket(
    ticket_data: TicketCreate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """Raise a support ticket for one of your own transactions."""
    return support_ticket_service.create_ticket(db, current_user, ticket_data)


@router.get("/", response_model=list[TicketResponse])
def list_my_tickets(
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """View all tickets you've raised."""
    return support_ticket_service.get_user_tickets(db, current_user)


@router.get("/{ticket_id}", response_model=TicketResponse)
def get_ticket(
    ticket_id: str,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """Check the status of a specific ticket you raised."""
    return support_ticket_service.get_ticket_by_id(db, current_user, ticket_id)


# ---------- Admin-only routes ----------

@router.get("/admin/all", response_model=list[TicketResponse])
def list_all_tickets(
    db: Session = Depends(get_db),
    admin: User = Depends(get_current_admin)
):
    """Admin: view every support ticket across all users."""
    return support_ticket_service.get_all_tickets(db)


@router.patch("/admin/{ticket_id}/status", response_model=TicketResponse)
def update_ticket_status(
    ticket_id: int,
    status_update: TicketStatusUpdate,
    db: Session = Depends(get_db),
    admin: User = Depends(get_current_admin)
):
    """Admin: update a ticket's status (OPEN -> IN_PROGRESS -> CLOSED), or close it directly."""
    return support_ticket_service.update_ticket_status(db, ticket_id, status_update.status)
