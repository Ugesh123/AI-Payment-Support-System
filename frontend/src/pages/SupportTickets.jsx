/**
 * Support ticket page.
 * Lists all tickets raised by the user, with a "New Ticket" button
 * that opens a form scoped to one of their own transactions.
 */
import { useEffect, useState } from "react";
import Card from "../components/ui/Card";
import Badge from "../components/ui/Badge";
import Loader from "../components/ui/Loader";
import ErrorMessage from "../components/ui/ErrorMessage";
import EmptyState from "../components/ui/EmptyState";
import Button from "../components/ui/Button";
import CreateTicketModal from "../components/CreateTicketModal";
import { getMyTickets } from "../services/ticketService";
import { formatDateTime } from "../utils/formatters";

export default function SupportTickets() {
  const [tickets, setTickets] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [showCreateModal, setShowCreateModal] = useState(false);

  const loadTickets = async () => {
    setLoading(true);
    try {
      const data = await getMyTickets();
      setTickets(data);
    } catch (err) {
      setError(err.response?.data?.detail || "Failed to load tickets.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadTickets();
  }, []);

  const handleCreated = () => {
    setShowCreateModal(false);
    loadTickets();
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-5">
        <h1 className="text-xl font-semibold text-gray-800">Support Tickets</h1>
        <Button onClick={() => setShowCreateModal(true)}>+ New Ticket</Button>
      </div>

      <Card>
        {loading ? (
          <Loader label="Loading tickets..." />
        ) : error ? (
          <ErrorMessage message={error} />
        ) : tickets.length === 0 ? (
          <EmptyState
            title="No support tickets yet"
            subtitle="Raise a ticket if you have an issue with a payment."
          />
        ) : (
          <div className="space-y-4">
            {tickets.map((ticket) => (
              <div key={ticket.ticket_id} className="border rounded-lg p-4">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <p className="font-mono text-xs text-gray-400">{ticket.ticket_id}</p>
                    <p className="font-medium text-gray-800 mt-0.5">{ticket.subject}</p>
                  </div>
                  <Badge status={ticket.status} />
                </div>

                <p className="text-sm text-gray-600 mt-2">{ticket.description}</p>

                <div className="flex justify-between text-xs text-gray-400 mt-3">
                  <span>Opened: {formatDateTime(ticket.created_at)}</span>
                  {ticket.closed_at && <span>Closed: {formatDateTime(ticket.closed_at)}</span>}
                </div>
              </div>
            ))}
          </div>
        )}
      </Card>

      <CreateTicketModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onCreated={handleCreated}
      />
    </div>
  );
}
