/**
 * Global toast notification system, following the exact same
 * Context pattern as AuthContext.jsx -- a provider at the app root,
 * and a useToast() hook any component can call to show a message.
 */
import { createContext, useContext, useState, useCallback } from "react";

const ToastContext = createContext(null);

export function ToastProvider({ children }) {
  const [toast, setToast] = useState(null); // { message, type }

  const showToast = useCallback((message, type = "success") => {
    setToast({ message, type });
    setTimeout(() => setToast(null), 3000);
  }, []);

  return (
    <ToastContext.Provider value={{ showToast }}>
      {children}
      {toast && <ToastBanner message={toast.message} type={toast.type} />}
    </ToastContext.Provider>
  );
}

export function useToast() {
  return useContext(ToastContext);
}

function ToastBanner({ message, type }) {
  const styles = {
    success: "bg-green-600 text-white",
    error: "bg-red-600 text-white",
  };

  return (
    <div className="fixed bottom-5 right-5 z-50">
      <div className={`px-4 py-3 rounded-lg shadow-lg text-sm ${styles[type]}`}>
        {message}
      </div>
    </div>
  );
}
