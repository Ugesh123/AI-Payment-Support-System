/**
 * Global authentication state.
 * On app load, if a token exists in localStorage, it's verified
 * against GET /auth/me. If invalid/expired, the user is logged out
 * automatically instead of the app silently breaking on every request.
 */
import { createContext, useContext, useState, useEffect } from "react";
import axiosClient from "../api/axiosClient";

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [token, setToken] = useState(localStorage.getItem("access_token"));
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true); // true while verifying token on load

  useEffect(() => {
    const verifyToken = async () => {
      const storedToken = localStorage.getItem("access_token");

      if (!storedToken) {
        setLoading(false);
        return;
      }

      try {
        const response = await axiosClient.get("/auth/me");
        setUser(response.data);
        setToken(storedToken);
      } catch (err) {
        localStorage.removeItem("access_token");
        setToken(null);
        setUser(null);
      } finally {
        setLoading(false);
      }
    };

    verifyToken();
  }, []);

  const login = async (newToken) => {
    localStorage.setItem("access_token", newToken);
    setToken(newToken);
    const response = await axiosClient.get("/auth/me");
    setUser(response.data);
  };

  const logout = () => {
    localStorage.removeItem("access_token");
    setToken(null);
    setUser(null);
  };

  return (
    <AuthContext.Provider
      value={{ token, user, isAuthenticated: !!token, loading, login, logout }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
