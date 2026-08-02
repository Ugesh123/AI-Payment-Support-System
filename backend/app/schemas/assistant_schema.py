"""
Pydantic schemas for the AI Assistant chat endpoint.
"""
from pydantic import BaseModel

class AssistantRequest(BaseModel):
    transaction_id: str
    question: str

class AssistantResponse(BaseModel):
    answer: str
