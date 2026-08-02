/**
 * Refund management page.
 * Lists the user's SUCCESS transactions (only these are refund-eligible)
 * and shows the refund status if one has been requested, or a
 * "Request Refund" button if not.
 */
import { useEffect, useState } from "react";
import Card from "../components/ui/Card";
import Badge from "../components/ui/Badge";
import Loader from "../components/ui/Loader";
import ErrorMessage from "../components/ui/ErrorMessage";
import EmptyState from "../components/ui/EmptyState";
import Button from "../components/ui/Button";
import RequestRefundModal from "../components/RequestRefundModal";
import { getTransactions } from "../services/transactionService";
import { getRefundStatus } from "../services/refundService";
import { formatCurrency, formatDateTime } from "../utils/formatters";

const EXPECTED_REFUND_DAYS = 3;

export default function Refunds() {
  const [transactions, setTransactions] = useState([]);
  const [refundsByTxn, setRefundsByTxn] = useState({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [selectedTxn, setSelectedTxn] = useState(null);

  const loadData = async () => {
    setLoading(true);
    try {
      const txns = await getTransactions();
      const successTxns = txns.filter((t) => t.status === "SUCCESS");
      setTransactions(successTxns);

      const refundMap = {};
      await Promise.all(
        successTxns.map(async (txn) => {
          try {
            const refund = await getRefundStatus(txn.transaction_id);
            refundMap[txn.transaction_id] = refund;
          } catch {
            // no refund exists for this transaction -- expected, not an error
          }
        })
      );
      setRefundsByTxn(refundMap);
    } catch (err) {
      setError(err.response?.data?.detail || "Failed to load refunds.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleRequested = () => {
    setSelectedTxn(null);
    loadData();
  };

  return (
    <div>
      <h1 className="text-xl font-semibold text-gray-800 mb-5">Refunds</h1>

      <Card>
        {loading ? (
          <Loader label="Loading refund-eligible transactions..." />
        ) : error ? (
          <ErrorMessage message={error} />
        ) : transactions.length === 0 ? (
          <EmptyState
            title="No refund-eligible transactions"
            subtitle="Only successful payments can be refunded."
          />
        ) : (
          <div className="space-y-4">
            {transactions.map((txn) => {
              const refund = refundsByTxn[txn.transaction_id];
              return (
                <div
                  key={txn.transaction_id}
                  className="border rounded-lg p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3"
                >
                  <div>
                    <p className="font-mono text-sm text-gray-700">{txn.transaction_id}</p>
                    <p className="text-sm text-gray-500">
                      {txn.receiver_name} • {formatCurrency(txn.amount)}
                    </p>
                  </div>

                  {refund ? (
                    <RefundTimeline refund={refund} />
                  ) : (
                    <Button onClick={() => setSelectedTxn(txn)}>Request Refund</Button>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </Card>

      <RequestRefundModal
        transaction={selectedTxn}
        isOpen={!!selectedTxn}
        onClose={() => setSelectedTxn(null)}
        onRequested={handleRequested}
      />
    </div>
  );
}

function RefundTimeline({ refund }) {
  return (
    <div className="text-sm text-right sm:text-left">
      <Badge status={refund.status} />
      <p className="text-gray-500 mt-1">
        Requested: {formatDateTime(refund.requested_at)}
      </p>
      {refund.status === "COMPLETED" ? (
        <p className="text-gray-500">Completed: {formatDateTime(refund.completed_at)}</p>
      ) : (
        <p className="text-gray-500">
          Expected: within {EXPECTED_REFUND_DAYS} business days
        </p>
      )}
    </div>
  );
}
