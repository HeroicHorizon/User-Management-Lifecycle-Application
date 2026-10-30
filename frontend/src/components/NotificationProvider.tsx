import { createContext, useCallback, useContext, useState, type ReactNode } from "react";

type Kind = "success" | "error";
interface Toast {
  id: number;
  kind: Kind;
  message: string;
}

interface NotifyApi {
  success: (message: string) => void;
  error: (message: string) => void;
}

const NotificationContext = createContext<NotifyApi | null>(null);
let nextId = 1;

export function NotificationProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const push = useCallback((kind: Kind, message: string) => {
    const id = nextId++;
    setToasts((t) => [...t, { id, kind, message }]);
    setTimeout(() => setToasts((t) => t.filter((x) => x.id !== id)), 6000);
  }, []);

  const api: NotifyApi = {
    success: useCallback((m: string) => push("success", m), [push]),
    error: useCallback((m: string) => push("error", m), [push]),
  };

  return (
    <NotificationContext.Provider value={api}>
      {children}
      <div className="toasts">
        {toasts.map((t) => (
          <div key={t.id} className={`toast toast-${t.kind}`} role={t.kind === "error" ? "alert" : "status"}>
            {t.message}
          </div>
        ))}
      </div>
    </NotificationContext.Provider>
  );
}

export function useNotify(): NotifyApi {
  const ctx = useContext(NotificationContext);
  if (!ctx) throw new Error("useNotify must be used inside NotificationProvider");
  return ctx;
}
