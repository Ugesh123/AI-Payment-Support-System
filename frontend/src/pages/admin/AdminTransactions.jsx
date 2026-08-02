/**
 * Admin: view all transactions, with confirmation before any status
 * change, both row buttons disabled during the update, and a toast
 * shown on success.
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
import { getAllTransactions, updateTransactionStatus } from "../../services/adminService";
import { formatCurrency, formatDateTime } from "../../utils/formatters";
import { useToast } from "../../context/ToastContext";

export default function AdminTransactions() {
  const [transactions, setTransactions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [updatingId, setUpdatingId] = useState(null);
  const [pendingAction, setPendingAction] = useState(null);
  const { showToast } = useToast();

  const load = async () => {
    try {
      setTransactions(await getAllTransactions());
    } catch (err) {
      setError(err.response?.data?.detail || "Failed to load transactions.");
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
      await updateTransactionStatus(id, status);
      showToast(`Transaction marked as ${status}.`);
      await load();
    } catch (err) {
      showToast(err.response?.data?.detail || "Failed to update transaction.", "error");
    } finally {
      setUpdatingId(null);
      setPendingAction(null);
    }
  };

  return (
    <div>
      <h1 className="text-xl font-semibold text-gray-800 mb-5">All Transactions</h1>
      <AdminTabs />

      <Card>
        {loading ? (
          <Loader label="Loading transactions..." />
        ) : error ? (
          <ErrorMessage message={error} />
        ) : transactions.length === 0 ? (
          <EmptyState title="No transactions found" />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-gray-500 border-b">
                  <th className="py-2 pr-4">Transaction ID</th>
                  <th className="py-2 pr-4">Amount</th>
                  <th className="py-2 pr-4">Status</th>
                  <th className="py-2 pr-4">Date</th>
                  <th className="py-2 pr-4">Action</th>
                </tr>
              </thead>
              <tbody>
                {transactions.map((txn) => {
                  const isUpdatingThisRow = updatingId === txn.id;
                  return (
                    <tr key={txn.id} className="border-b last:border-0">
                      <td className="py-3 pr-4 font-mono text-gray-700">{txn.transaction_id}</td>
                      <td className="py-3 pr-4 text-gray-700">{formatCurrency(txn.amount)}</td>
                      <td className="py-3 pr-4"><Badge status={txn.status} /></td>
                      <td className="py-3 pr-4 text-gray-500">{formatDateTime(txn.created_at)}</td>
                      <td className="py-3 pr-4">
                        {txn.status === "PENDING" ? (
                          <div className="flex gap-2">
                            <Button
                              variant="secondary"
                              disabled={isUpdatingThisRow}
                              onClick={() => requestUpdate(txn.id, "SUCCESS")}
                            >
                              Mark Success
                            </Button>
                            <Button
                              variant="danger"
                              disabled={isUpdatingThisRow}
                              onClick={() => requestUpdate(txn.id, "FAILED")}
                            >
                              Mark Failed
                            </Button>
                          </div>
                        ) : (
                          <span className="text-gray-400 text-xs">No action</span>
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
        title="Confirm Status Change"
        message={
          pendingAction
            ? `Mark this transaction as ${pendingAction.status}? This cannot be undone.`
            : ""
        }
        onConfirm={confirmUpdate}
        onCancel={() => setPendingAction(null)}
        loading={!!updatingId}
      />
    </div>
  );
}
