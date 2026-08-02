/**
 * Transaction history page.
 * Includes client-side search (by transaction ID), status filter,
 * pagination, and a "Download Receipt" action for SUCCESS transactions.
 */
import { useEffect, useState, useMemo } from "react";
import Card from "../components/ui/Card";
import Badge from "../components/ui/Badge";
import Loader from "../components/ui/Loader";
import ErrorMessage from "../components/ui/ErrorMessage";
import EmptyState from "../components/ui/EmptyState";
import Button from "../components/ui/Button";
import CreateTransactionModal from "../components/CreateTransactionModal";
import { getTransactions } from "../services/transactionService";
import { downloadReceipt } from "../services/receiptService";
import { downloadFile } from "../utils/downloadFile";
import { formatCurrency, formatDateTime } from "../utils/formatters";
import { useToast } from "../context/ToastContext";

const PAGE_SIZE = 8;
const STATUS_OPTIONS = ["ALL", "SUCCESS", "FAILED", "PENDING"];

export default function Transactions() {
  const [transactions, setTransactions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("ALL");
  const [page, setPage] = useState(1);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [downloadingId, setDownloadingId] = useState(null);
  const { showToast } = useToast();

  const loadTransactions = async () => {
    setLoading(true);
    try {
      const data = await getTransactions();
      setTransactions(data);
    } catch (err) {
      setError(err.response?.data?.detail || "Failed to load transactions.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadTransactions();
  }, []);

  const filtered = useMemo(() => {
    return transactions.filter((txn) => {
      const matchesStatus = statusFilter === "ALL" || txn.status === statusFilter;
      const matchesSearch = txn.transaction_id
        .toLowerCase()
        .includes(search.toLowerCase());
      return matchesStatus && matchesSearch;
    });
  }, [transactions, search, statusFilter]);

  const totalPages = Math.max(1, Math.ceil(filtered.length / PAGE_SIZE));
  const paginated = filtered.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE);

  useEffect(() => {
    setPage(1);
  }, [search, statusFilter]);

  const handleCreated = () => {
    setShowCreateModal(false);
    loadTransactions();
  };

  const handleDownloadReceipt = async (transactionId) => {
    setDownloadingId(transactionId);
    try {
      const pdfBlob = await downloadReceipt(transactionId);
      downloadFile(pdfBlob, `receipt_${transactionId}.pdf`);
      showToast("Receipt downloaded successfully.");
    } catch (err) {
      showToast(
        err.response?.data?.detail || "Receipt can only be generated for successful transactions.",
        "error"
      );
    } finally {
      setDownloadingId(null);
    }
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-5">
        <h1 className="text-xl font-semibold text-gray-800">Transactions</h1>
        <Button onClick={() => setShowCreateModal(true)}>+ New Payment</Button>
      </div>

      <div className="flex flex-col sm:flex-row gap-3 mb-4">
        <input
          type="text"
          placeholder="Search by Transaction ID..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="border rounded px-3 py-2 text-sm w-full sm:w-64 focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="border rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          {STATUS_OPTIONS.map((s) => (
            <option key={s} value={s}>
              {s === "ALL" ? "All Statuses" : s}
            </option>
          ))}
        </select>
      </div>

      <Card>
        {loading ? (
          <Loader label="Loading transactions..." />
        ) : error ? (
          <ErrorMessage message={error} />
        ) : filtered.length === 0 ? (
          <EmptyState
            title="No transactions found"
            subtitle="Try adjusting your search or filter."
          />
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="text-left text-gray-500 border-b">
                    <th className="py-2 pr-4">Transaction ID</th>
                    <th className="py-2 pr-4">Receiver</th>
                    <th className="py-2 pr-4">Amount</th>
                    <th className="py-2 pr-4">Method</th>
                    <th className="py-2 pr-4">Status</th>
                    <th className="py-2 pr-4">Date</th>
                    <th className="py-2 pr-4">Receipt</th>
                  </tr>
                </thead>
                <tbody>
                  {paginated.map((txn) => (
                    <tr key={txn.transaction_id} className="border-b last:border-0">
                      <td className="py-3 pr-4 font-mono text-gray-700">{txn.transaction_id}</td>
                      <td className="py-3 pr-4 text-gray-700">{txn.receiver_name}</td>
                      <td className="py-3 pr-4 text-gray-700">{formatCurrency(txn.amount)}</td>
                      <td className="py-3 pr-4 text-gray-700">{txn.payment_method}</td>
                      <td className="py-3 pr-4">
                        <Badge status={txn.status} />
                        {txn.status === "FAILED" && txn.failure_reason && (
                          <p className="text-xs text-red-500 mt-1">{txn.failure_reason}</p>
                        )}
                      </td>
                      <td className="py-3 pr-4 text-gray-500">{formatDateTime(txn.created_at)}</td>
                      <td className="py-3 pr-4">
                        {txn.status === "SUCCESS" ? (
                          <Button
                            variant="outline"
                            loading={downloadingId === txn.transaction_id}
                            onClick={() => handleDownloadReceipt(txn.transaction_id)}
                          >
                            Download
                          </Button>
                        ) : (
                          <span className="text-gray-400 text-xs">Not available</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div className="flex items-center justify-between mt-4 text-sm">
              <span className="text-gray-500">
                Page {page} of {totalPages}
              </span>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  disabled={page === 1}
                  onClick={() => setPage((p) => p - 1)}
                >
                  Previous
                </Button>
                <Button
                  variant="outline"
                  disabled={page === totalPages}
                  onClick={() => setPage((p) => p + 1)}
                >
                  Next
                </Button>
              </div>
            </div>
          </>
        )}
      </Card>

      <CreateTransactionModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onCreated={handleCreated}
      />
    </div>
  );
}
