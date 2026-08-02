/**
 * User dashboard page.
 * Fetches summary cards + recent transactions in a single API call
 * (GET /dashboard/), matching the backend's combined response shape.
 */
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import Card from "../components/ui/Card";
import Badge from "../components/ui/Badge";
import Loader from "../components/ui/Loader";
import ErrorMessage from "../components/ui/ErrorMessage";
import EmptyState from "../components/ui/EmptyState";
import { getDashboard } from "../services/dashboardService";
import { formatCurrency } from "../utils/formatters";

const SUMMARY_CARDS = [
  { key: "total_transactions", label: "Total Transactions" },
  { key: "successful_payments", label: "Successful Payments" },
  { key: "failed_payments", label: "Failed Payments" },
  { key: "pending_payments", label: "Pending Payments" },
  { key: "total_amount_paid", label: "Total Amount Paid", isCurrency: true },
  { key: "active_refunds", label: "Active Refunds" },
  { key: "completed_refunds", label: "Completed Refunds" },
];

export default function Dashboard() {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    const fetchDashboard = async () => {
      try {
        const result = await getDashboard();
        setData(result);
      } catch (err) {
        setError(err.response?.data?.detail || "Failed to load dashboard.");
      } finally {
        setLoading(false);
      }
    };
    fetchDashboard();
  }, []);

  if (loading) return <Loader label="Loading dashboard..." />;
  if (error) return <ErrorMessage message={error} />;

  const { summary, recent_transactions } = data;

  return (
    <div>
      <h1 className="text-xl font-semibold text-gray-800 mb-5">Dashboard</h1>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        {SUMMARY_CARDS.map((card) => (
          <Card key={card.key}>
            <p className="text-sm text-gray-500">{card.label}</p>
            <p className="text-2xl font-semibold text-gray-800 mt-1">
              {card.isCurrency ? formatCurrency(summary[card.key]) : summary[card.key]}
            </p>
          </Card>
        ))}
      </div>

      <Card>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-base font-semibold text-gray-800">Recent Transactions</h2>
          <Link to="/transactions" className="text-sm text-blue-600 hover:underline">
            View all
          </Link>
        </div>

        {recent_transactions.length === 0 ? (
          <EmptyState
            title="No transactions yet"
            subtitle="Your recent payments will show up here."
          />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-gray-500 border-b">
                  <th className="py-2 pr-4">Transaction ID</th>
                  <th className="py-2 pr-4">Receiver</th>
                  <th className="py-2 pr-4">Amount</th>
                  <th className="py-2 pr-4">Status</th>
                </tr>
              </thead>
              <tbody>
                {recent_transactions.map((txn) => (
                  <tr key={txn.transaction_id} className="border-b last:border-0">
                    <td className="py-3 pr-4 font-mono text-gray-700">{txn.transaction_id}</td>
                    <td className="py-3 pr-4 text-gray-700">{txn.receiver}</td>
                    <td className="py-3 pr-4 text-gray-700">{formatCurrency(txn.amount)}</td>
                    <td className="py-3 pr-4">
                      <Badge status={txn.status} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>
    </div>
  );
}
