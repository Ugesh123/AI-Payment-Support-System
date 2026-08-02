"""
API endpoint for the AI Assistant.
"""
from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session
from app.database.db import get_db
from app.schemas.assistant_schema import AssistantRequest, AssistantResponse
from app.services import ai_service
from app.services.dependencies import get_current_user
from app.models.user import User

router = APIRouter(prefix="/assistant", tags=["AI Assistant"])

@router.post("/chat", response_model=AssistantResponse)
def chat_with_assistant(
    request: AssistantRequest,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """Answers a question about a specific transaction, grounded in real MySQL data."""
    answer = ai_service.ask_assistant(db, current_user, request.transaction_id, request.question)
    return AssistantResponse(answer=answer)
