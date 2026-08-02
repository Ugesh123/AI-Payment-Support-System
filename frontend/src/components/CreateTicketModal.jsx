/**
 * "New Ticket" form, shown inside the shared Modal.
 * Ticket must be tied to one of the user's own transactions.
 */
import { useState, useEffect } from "react";
import Modal from "./ui/Modal";
import Button from "./ui/Button";
import ErrorMessage from "./ui/ErrorMessage";
import { getTransactions } from "../services/transactionService";
import { createTicket } from "../services/ticketService";

export default function CreateTicketModal({ isOpen, onClose, onCreated }) {
  const [transactions, setTransactions] = useState([]);
  const [form, setForm] = useState({ transaction_id: "", subject: "", description: "" });
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!isOpen) return;
    const loadTransactions = async () => {
      try {
        const data = await getTransactions();
        setTransactions(data);
        if (data.length > 0) {
          setForm((prev) => ({ ...prev, transaction_id: data[0].transaction_id }));
        }
      } catch {
        setError("Failed to load your transactions.");
      }
    };
    loadTransactions();
  }, [isOpen]);

  const handleChange = (e) => {
    setForm({ ...form, [e.target.name]: e.target.value });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    setSubmitting(true);
    try {
      await createTicket(form);
      setForm({ transaction_id: "", subject: "", description: "" });
      onCreated();
    } catch (err) {
      setError(err.response?.data?.detail || "Failed to create ticket.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Raise a Support Ticket">
      <form onSubmit={handleSubmit}>
        {error && <div className="mb-3"><ErrorMessage message={error} /></div>}

        {transactions.length === 0 ? (
          <p className="text-sm text-gray-500 mb-4">
            You need at least one transaction before raising a ticket.
          </p>
        ) : (
          <>
            <label className="block text-sm text-gray-600 mb-1">Related Transaction</label>
            <select
              name="transaction_id"
              value={form.transaction_id}
              onChange={handleChange}
              className="w-full border rounded px-3 py-2 mb-3 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              {transactions.map((txn) => (
                <option key={txn.transaction_id} value={txn.transaction_id}>
                  {txn.transaction_id} -- {txn.status}
                </option>
              ))}
            </select>

            <label className="block text-sm text-gray-600 mb-1">Subject</label>
            <input
              type="text"
              name="subject"
              value={form.subject}
              onChange={handleChange}
              required
              placeholder="e.g. Payment deducted but order not confirmed"
              className="w-full border rounded px-3 py-2 mb-3 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />

            <label className="block text-sm text-gray-600 mb-1">Description</label>
            <textarea
              name="description"
              value={form.description}
              onChange={handleChange}
              required
              rows={4}
              placeholder="Describe your issue in detail..."
              className="w-full border rounded px-3 py-2 mb-4 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </>
        )}

        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={onClose}>Cancel</Button>
          <Button type="submit" loading={submitting} disabled={transactions.length === 0}>
            Submit Ticket
          </Button>
        </div>
      </form>
    </Modal>
  );
}
