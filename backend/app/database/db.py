"""
Sets up the SQLAlchemy engine and session.
Every route/service that needs DB access will use `get_db()`.
"""
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker, declarative_base
from app.config import settings

engine = create_engine(settings.DATABASE_URL)
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

Base = declarative_base()

def get_db():
    """
    Dependency used by FastAPI routes.
    Opens a DB session, yields it, and always closes it after the request.
    """
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()
