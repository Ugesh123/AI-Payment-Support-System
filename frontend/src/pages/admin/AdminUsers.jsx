/**
 * Admin: view all registered users. Read-only.
 */
import { useEffect, useState } from "react";
import Card from "../../components/ui/Card";
import Badge from "../../components/ui/Badge";
import Loader from "../../components/ui/Loader";
import ErrorMessage from "../../components/ui/ErrorMessage";
import EmptyState from "../../components/ui/EmptyState";
import AdminTabs from "../../components/admin/AdminTabs";
import { getAllUsers } from "../../services/adminService";
import { formatDateTime } from "../../utils/formatters";

export default function AdminUsers() {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    const load = async () => {
      try {
        setUsers(await getAllUsers());
      } catch (err) {
        setError(err.response?.data?.detail || "Failed to load users.");
      } finally {
        setLoading(false);
      }
    };
    load();
  }, []);

  return (
    <div>
      <h1 className="text-xl font-semibold text-gray-800 mb-5">All Users</h1>
      <AdminTabs />

      <Card>
        {loading ? (
          <Loader label="Loading users..." />
        ) : error ? (
          <ErrorMessage message={error} />
        ) : users.length === 0 ? (
          <EmptyState title="No users found" />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-gray-500 border-b">
                  <th className="py-2 pr-4">Name</th>
                  <th className="py-2 pr-4">Email</th>
                  <th className="py-2 pr-4">Role</th>
                  <th className="py-2 pr-4">Joined</th>
                </tr>
              </thead>
              <tbody>
                {users.map((u) => (
                  <tr key={u.id} className="border-b last:border-0">
                    <td className="py-3 pr-4 text-gray-700">{u.name}</td>
                    <td className="py-3 pr-4 text-gray-700">{u.email}</td>
                    <td className="py-3 pr-4">
                      <Badge status={u.role.toUpperCase()} />
                    </td>
                    <td className="py-3 pr-4 text-gray-500">{formatDateTime(u.created_at)}</td>
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
