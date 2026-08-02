/**
 * Admin: view all refund requests. Same confirm-then-toast pattern
 * as AdminTransactions.
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
import { getAllRefunds, updateRefundStatus } from "../../services/adminService";
import { formatCurrency, formatDateTime } from "../../utils/formatters";
import { useToast } from "../../context/ToastContext";

const NEXT_STATUS = {
  REQUESTED: "PROCESSING",
  PROCESSING: "COMPLETED",
};

export default function AdminRefunds() {
  const [refunds, setRefunds] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [updatingId, setUpdatingId] = useState(null);
  const [pendingAction, setPendingAction] = useState(null);
  const { showToast } = useToast();

  const load = async () => {
    try {
      setRefunds(await getAllRefunds());
    } catch (err) {
      setError(err.response?.data?.detail || "Failed to load refunds.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const requestAdvance = (id, status) => {
    setPendingAction({ id, status });
  };

  const confirmAdvance = async () => {
    if (!pendingAction) return;
    const { id, status } = pendingAction;
    setUpdatingId(id);
    try {
      await updateRefundStatus(id, status);
      showToast(`Refund moved to ${status}.`);
      await load();
    } catch (err) {
      showToast(err.response?.data?.detail || "Failed to update refund.", "error");
    } finally {
      setUpdatingId(null);
      setPendingAction(null);
    }
  };

  return (
    <div>
      <h1 className="text-xl font-semibold text-gray-800 mb-5">All Refund Requests</h1>
      <AdminTabs />

      <Card>
        {loading ? (
          <Loader label="Loading refunds..." />
        ) : error ? (
          <ErrorMessage message={error} />
        ) : refunds.length === 0 ? (
          <EmptyState title="No refund requests found" />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-gray-500 border-b">
                  <th className="py-2 pr-4">Amount</th>
                  <th className="py-2 pr-4">Reason</th>
                  <th className="py-2 pr-4">Status</th>
                  <th className="py-2 pr-4">Requested</th>
                  <th className="py-2 pr-4">Action</th>
                </tr>
              </thead>
              <tbody>
                {refunds.map((r) => {
                  const next = NEXT_STATUS[r.status];
                  const isUpdatingThisRow = updatingId === r.id;
                  return (
                    <tr key={r.id} className="border-b last:border-0">
                      <td className="py-3 pr-4 text-gray-700">{formatCurrency(r.amount)}</td>
                      <td className="py-3 pr-4 text-gray-700">{r.reason}</td>
                      <td className="py-3 pr-4"><Badge status={r.status} /></td>
                      <td className="py-3 pr-4 text-gray-500">{formatDateTime(r.requested_at)}</td>
                      <td className="py-3 pr-4">
                        {next ? (
                          <Button
                            variant="secondary"
                            disabled={isUpdatingThisRow}
                            onClick={() => requestAdvance(r.id, next)}
                          >
                            Move to {next}
                          </Button>
                        ) : (
                          <span className="text-gray-400 text-xs">Completed</span>
                        )}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      <ConfirmModal
        isOpen={!!pendingAction}
        title="Confirm Refund Update"
        message={
          pendingAction ? `Move this refund to ${pendingAction.status}?` : ""
        }
        onConfirm={confirmAdvance}
        onCancel={() => setPendingAction(null)}
        loading={!!updatingId}
      />
    </div>
  );
}
