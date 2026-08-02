"""
API endpoint for downloading a transaction receipt as a PDF.
"""
from fastapi import APIRouter, Depends
from fastapi.responses import Response
from sqlalchemy.orm import Session
from app.database.db import get_db
from app.services import receipt_service
from app.services.dependencies import get_current_user
from app.models.user import User

router = APIRouter(prefix="/receipt", tags=["Receipt"])


@router.get("/{transaction_id}")
def download_receipt(
    transaction_id: str,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Downloads the PDF receipt for a successful transaction.
    Returns 400 if the transaction isn't SUCCESS, 404 if it doesn't
    belong to the user.
    """
    pdf_bytes = receipt_service.generate_receipt(db, current_user, transaction_id)

    return Response(
        content=pdf_bytes,
        media_type="application/pdf",
        headers={
            "Content-Disposition": f'attachment; filename="receipt_{transaction_id}.pdf"',
            "Cache-Control": "no-store"
        }
    )
