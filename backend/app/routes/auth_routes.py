"""
API endpoints for authentication.
"""
from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session
from app.database.db import get_db
from app.schemas.user_schema import UserRegister, UserLogin, UserResponse, Token
from app.services import auth_service
from app.services.dependencies import get_current_user
from app.models.user import User

router = APIRouter(prefix="/auth", tags=["Authentication"])

@router.post("/register", response_model=UserResponse)
def register(user_data: UserRegister, db: Session = Depends(get_db)):
    return auth_service.register_user(db, user_data)

@router.post("/login", response_model=Token)
def login(login_data: UserLogin, db: Session = Depends(get_db)):
    token = auth_service.login_user(db, login_data)
    return Token(access_token=token)

@router.get("/me", response_model=UserResponse)
def get_me(current_user: User = Depends(get_current_user)):
    """
    Returns the currently logged-in user's info.
    Used by the frontend on app load to verify the token is still
    valid and to fetch user details (name, email, role) for the UI.
    """
    return current_user
