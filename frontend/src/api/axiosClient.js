/**
 * Pre-configured Axios instance.
 * Automatically attaches the JWT token (if present) to every request,
 * so individual pages/components don't need to manage headers manually.
 */
import axios from "axios";

const axiosClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
});

axiosClient.interceptors.request.use((config) => {
  const token = localStorage.getItem("access_token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export default axiosClient;
