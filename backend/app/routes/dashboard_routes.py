"""
API endpoint for the dashboard summary.
"""
from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session
from app.database.db import get_db
from app.schemas.dashboard_schema import DashboardResponse
from app.services import dashboard_service
from app.services.dependencies import get_current_user
from app.models.user import User

router = APIRouter(prefix="/dashboard", tags=["Dashboard"])

@router.get("/", response_model=DashboardResponse)
def get_dashboard(
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """Returns summary cards + the 5 most recent transactions."""
    return dashboard_service.get_dashboard(db, current_user)
