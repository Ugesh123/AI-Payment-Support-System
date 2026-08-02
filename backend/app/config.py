"""
Centralized app settings.
Loads secrets and configuration from environment variables instead of
hardcoding them, so the same code works locally and after deployment.
"""
import os
from dotenv import load_dotenv

load_dotenv()

class Settings:
    DATABASE_URL: str = os.getenv("DATABASE_URL")
    JWT_SECRET_KEY: str = os.getenv("JWT_SECRET_KEY")
    JWT_ALGORITHM: str = os.getenv("JWT_ALGORITHM", "HS256")
    ACCESS_TOKEN_EXPIRE_MINUTES: int = int(os.getenv("ACCESS_TOKEN_EXPIRE_MINUTES", 60))
    GEMINI_API_KEY: str = os.getenv("GEMINI_API_KEY")

settings = Settings()
