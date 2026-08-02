"""
Receipt generation service.
Builds a simple PDF receipt in-memory using ReportLab and returns it
as raw bytes. Nothing is saved to disk -- the PDF is generated fresh
on every request directly from the transaction data in MySQL.
"""
import io
from reportlab.lib.pagesizes import A4
from reportlab.lib.units import cm
from reportlab.pdfgen import canvas
from sqlalchemy.orm import Session
from fastapi import HTTPException, status as http_status
from app.models.transaction import Transaction, TransactionStatus
from app.models.user import User


def _build_receipt_pdf(transaction: Transaction) -> bytes:
    """
    Draws a simple, clean receipt onto a PDF canvas and returns the
    raw bytes. Kept intentionally plain -- no logo, no QR code, no
    invoice numbering, per project scope.
    """
    buffer = io.BytesIO()
    pdf = canvas.Canvas(buffer, pagesize=A4)
    width, height = A4

    y = height - 3 * cm

    pdf.setFont("Helvetica-Bold", 16)
    pdf.drawString(2 * cm, y, "AI Payment Support System")
    y -= 0.8 * cm

    pdf.setFont("Helvetica", 11)
    pdf.drawString(2 * cm, y, "Payment Receipt")
    y -= 1.2 * cm

    pdf.line(2 * cm, y, width - 2 * cm, y)
    y -= 1 * cm

    fields = [
        ("Transaction ID", transaction.transaction_id),
        ("Date & Time", transaction.created_at.strftime("%Y-%m-%d %I:%M %p")),
        ("Amount", f"Rs. {transaction.amount}"),
        ("Payment Method", transaction.payment_method.value),
        ("Status", transaction.status.value),
        ("Receiver Name", transaction.receiver_name),
    ]

    pdf.setFont("Helvetica", 12)
    for label, value in fields:
        pdf.drawString(2 * cm, y, f"{label}:")
        pdf.drawString(7 * cm, y, str(value))
        y -= 0.9 * cm

    y -= 0.5 * cm
    pdf.line(2 * cm, y, width - 2 * cm, y)
    y -= 1 * cm

    pdf.setFont("Helvetica-Oblique", 10)
    pdf.drawString(2 * cm, y, "Thank you for using our application.")

    pdf.save()
    buffer.seek(0)
    return buffer.getvalue()


def generate_receipt(db: Session, user: User, transaction_id: str) -> bytes:
    """
    Main entry point. Fetches the transaction, enforces ownership and
    the SUCCESS-only business rule, then builds the PDF.
    """
    transaction = db.query(Transaction).filter(
        Transaction.transaction_id == transaction_id,
        Transaction.user_id == user.id
    ).first()

    if not transaction:
        raise HTTPException(status_code=http_status.HTTP_404_NOT_FOUND, detail="Transaction not found")

    if transaction.status != TransactionStatus.SUCCESS:
        raise HTTPException(
            status_code=http_status.HTTP_400_BAD_REQUEST,
            detail="Receipt can only be generated for successful transactions."
        )

    return _build_receipt_pdf(transaction)
