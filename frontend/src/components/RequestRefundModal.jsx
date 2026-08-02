/**
 * "Request Refund" form, shown inside a Modal.
 * Reason is a fixed dropdown, matching the backend's RefundReason enum.
 */
import { useState } from "react";
import Modal from "./ui/Modal";
import Button from "./ui/Button";
import ErrorMessage from "./ui/ErrorMessage";
import { requestRefund } from "../services/refundService";

const REFUND_REASONS = [
  "Duplicate Payment",
  "Wrong Amount",
  "Order Cancelled",
  "Other",
];

export default function RequestRefundModal({ transaction, isOpen, onClose, onRequested }) {
  const [reason, setReason] = useState(REFUND_REASONS[0]);
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    setSubmitting(true);
    try {
      await requestRefund({
        transaction_id: transaction.transaction_id,
        refund_reason: reason,
      });
      onRequested();
    } catch (err) {
      setError(err.response?.data?.detail || "Failed to request refund.");
    } finally {
      setSubmitting(false);
    }
  };

  if (!transaction) return null;

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Request Refund">
      <form onSubmit={handleSubmit}>
        {error && <div className="mb-3"><ErrorMessage message={error} /></div>}

        <p className="text-sm text-gray-600 mb-3">
          Transaction <span className="font-mono">{transaction.transaction_id}</span> --{" "}
          <span className="font-medium">₹{transaction.amount}</span>
        </p>

        <label className="block text-sm text-gray-600 mb-1">Reason</label>
        <select
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          className="w-full border rounded px-3 py-2 mb-4 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          {REFUND_REASONS.map((r) => (
            <option key={r} value={r}>{r}</option>
          ))}
        </select>

        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={onClose}>Cancel</Button>
          <Button type="submit" loading={submitting}>Submit Request</Button>
        </div>
      </form>
    </Modal>
  );
}
