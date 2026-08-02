/**
 * Admin: view all support tickets, with buttons to advance status.
 */
import { useEffect, useState } from "react";
import Card from "../../components/ui/Card";
import Badge from "../../components/ui/Badge";
import Loader from "../../components/ui/Loader";
import ErrorMessage from "../../components/ui/ErrorMessage";
import EmptyState from "../../components/ui/EmptyState";
import Button from "../../components/ui/Button";
import ConfirmModal from "../../components/ui/ConfirmModal";
import AdminTabs from "../../components/admin/AdminTabs";
import { getAllTickets, updateTicketStatus } from "../../services/adminService";
import { formatDateTime } from "../../utils/formatters";
import { useToast } from "../../context/ToastContext";

const NEXT_OPTIONS = {
  OPEN: ["IN_PROGRESS", "CLOSED"],
  IN_PROGRESS: ["CLOSED"],
};

export default function AdminTickets() {
  const [tickets, setTickets] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [updatingId, setUpdatingId] = useState(null);
  const [pendingAction, setPendingAction] = useState(null);
  const { showToast } = useToast();

  const load = async () => {
    try {
      setTickets(await getAllTickets());
    } catch (err) {
      setError(err.response?.data?.detail || "Failed to load tickets.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const requestUpdate = (id, status) => {
    setPendingAction({ id, status });
  };

  const confirmUpdate = async () => {
    if (!pendingAction) return;
    const { id, status } = pendingAction;
    setUpdatingId(id);
    try {
      await updateTicketStatus(id, status);
      showToast(
        status === "CLOSED" ? "Ticket closed successfully." : "Ticket marked in progress."
      );
      await load();
    } catch (err) {
      showToast(err.response?.data?.detail || "Failed to update ticket.", "error");
    } finally {
      setUpdatingId(null);
      setPendingAction(null);
    }
  };

  return (
    <div>
      <h1 className="text-xl font-semibold text-gray-800 mb-5">All Support Tickets</h1>
      <AdminTabs />

      <Card>
        {loading ? (
          <Loader label="Loading tickets..." />
        ) : error ? (
          <ErrorMessage message={error} />
        ) : tickets.length === 0 ? (
          <EmptyState title="No support tickets found" />
        ) : (
          <div className="space-y-4">
            {tickets.map((t) => {
              const options = NEXT_OPTIONS[t.status] || [];
              const isUpdatingThisRow = updatingId === t.id;
              return (
                <div key={t.id} className="border rounded-lg p-4">
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <p className="font-mono text-xs text-gray-400">{t.ticket_id}</p>
                      <p className="font-medium text-gray-800 mt-0.5">{t.subject}</p>
                      <p className="text-sm text-gray-600 mt-1">{t.description}</p>
                    </div>
                    <Badge status={t.status} />
                  </div>

                  <div className="flex items-center justify-between mt-3">
                    <span className="text-xs text-gray-400">
                      Opened: {formatDateTime(t.created_at)}
                    </span>
                    <div className="flex gap-2">
                      {options.map((status) => (
                        <Button
                          key={status}
                          variant={status === "CLOSED" ? "danger" : "secondary"}
                          disabled={isUpdatingThisRow}
                          onClick={() => requestUpdate(t.id, status)}
                        >
                          {status === "CLOSED" ? "Close Ticket" : "Mark In Progress"}
                        </Button>
                      ))}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </Card>

      <ConfirmModal
        isOpen={!!pendingAction}
        title="Confirm Ticket Update"
        message={
          pendingAction
            ? pendingAction.status === "CLOSED"
              ? "Close this ticket? This cannot be undone."
              : "Mark this ticket as in progress?"
            : ""
        }
        onConfirm={confirmUpdate}
        onCancel={() => setPendingAction(null)}
        loading={!!updatingId}
      />
    </div>
  );
}
