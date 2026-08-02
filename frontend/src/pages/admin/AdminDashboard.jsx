/**
 * Admin dashboard landing page -- single summary call.
 */
import { useEffect, useState } from "react";
import Card from "../../components/ui/Card";
import Loader from "../../components/ui/Loader";
import ErrorMessage from "../../components/ui/ErrorMessage";
import AdminTabs from "../../components/admin/AdminTabs";
import { getAdminDashboard } from "../../services/adminService";

const SUMMARY_CARDS = [
  { key: "total_users", label: "Total Users" },
  { key: "total_transactions", label: "Total Transactions" },
  { key: "pending_transactions", label: "Pending Transactions" },
  { key: "active_refunds", label: "Active Refunds" },
  { key: "open_tickets", label: "Open Tickets" },
];

export default function AdminDashboard() {
  const [summary, setSummary] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    const load = async () => {
      try {
        const data = await getAdminDashboard();
        setSummary(data);
      } catch (err) {
        setError(err.response?.data?.detail || "Failed to load admin dashboard.");
      } finally {
        setLoading(false);
      }
    };
    load();
  }, []);

  return (
    <div>
      <h1 className="text-xl font-semibold text-gray-800 mb-5">Admin Dashboard</h1>
      <AdminTabs />

      {loading ? (
        <Loader label="Loading admin dashboard..." />
      ) : error ? (
        <ErrorMessage message={error} />
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
          {SUMMARY_CARDS.map((card) => (
            <Card key={card.key}>
              <p className="text-sm text-gray-500">{card.label}</p>
              <p className="text-2xl font-semibold text-gray-800 mt-1">{summary[card.key]}</p>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
