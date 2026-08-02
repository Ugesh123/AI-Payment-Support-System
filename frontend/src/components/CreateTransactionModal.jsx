/**
 * "New Payment" form, shown inside a Modal.
 * After submitting, shows a second result modal telling the user
 * exactly what happened (SUCCESS/FAILED/PENDING).
 */
import { useState } from "react";
import Modal from "./ui/Modal";
import Button from "./ui/Button";
import Badge from "./ui/Badge";
import ErrorMessage from "./ui/ErrorMessage";
import { createTransaction } from "../services/transactionService";

const PAYMENT_METHODS = ["UPI", "Card", "Net Banking", "Wallet"];

const RESULT_MESSAGES = {
  SUCCESS: "Payment completed successfully.",
  FAILED: "Payment failed.",
  PENDING: "Payment is being processed.",
};

export default function CreateTransactionModal({ isOpen, onClose, onCreated }) {
  const [form, setForm] = useState({ amount: "", receiver_name: "", payment_method: "UPI" });
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [result, setResult] = useState(null);

  const handleChange = (e) => {
    setForm({ ...form, [e.target.name]: e.target.value });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    setSubmitting(true);
    try {
      const created = await createTransaction({ ...form, amount: parseFloat(form.amount) });
      setForm({ amount: "", receiver_name: "", payment_method: "UPI" });
      setResult(created);
    } catch (err) {
      setError(err.response?.data?.detail || "Failed to create transaction.");
    } finally {
      setSubmitting(false);
    }
  };

  const handleResultClose = () => {
    setResult(null);
    onCreated();
  };

  if (result) {
    return (
      <Modal isOpen={true} onClose={handleResultClose} title="Payment Result">
        <div className="text-center py-2">
          <Badge status={result.status} />
          <p className="text-gray-700 mt-3">{RESULT_MESSAGES[result.status]}</p>
          {result.status === "FAILED" && result.failure_reason && (
            <p className="text-sm text-red-500 mt-1">Reason: {result.failure_reason}</p>
          )}
          <p className="text-xs text-gray-400 mt-3 font-mono">{result.transaction_id}</p>
          <Button className="mt-5" onClick={handleResultClose}>Done</Button>
        </div>
      </Modal>
    );
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Make a Payment">
      <form onSubmit={handleSubmit}>
        {error && <div className="mb-3"><ErrorMessage message={error} /></div>}

        <label className="block text-sm text-gray-600 mb-1">Amount (₹)</label>
        <input
          type="number"
          name="amount"
          min="1"
          max="100000"
          step="0.01"
          value={form.amount}
          onChange={handleChange}
          required
          className="w-full border rounded px-3 py-2 mb-3 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        />

        <label className="block text-sm text-gray-600 mb-1">Receiver Name</label>
        <input
          type="text"
          name="receiver_name"
          value={form.receiver_name}
          onChange={handleChange}
          required
          placeholder="e.g. Amazon India"
          className="w-full border rounded px-3 py-2 mb-3 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        />

        <label className="block text-sm text-gray-600 mb-1">Payment Method</label>
        <select
          name="payment_method"
          value={form.payment_method}
          onChange={handleChange}
          className="w-full border rounded px-3 py-2 mb-4 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          {PAYMENT_METHODS.map((m) => (
            <option key={m} value={m}>{m}</option>
          ))}
        </select>

        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={onClose}>Cancel</Button>
          <Button type="submit" loading={submitting}>Pay Now</Button>
        </div>
      </form>
    </Modal>
  );
}
