import { createContext, useCallback, useState } from "react";
import { createPortal } from "react-dom";
import Toast, { ToastViewport, type ToastItem } from "@/components/ui/toast";
import type { ToastType } from "@/types";

type ToastContextType = {
  setters: {};
  data: {};
  actions: {
    showToast: (message: string, type?: ToastType) => void;
  };
};

export const ToastContext = createContext({} as ToastContextType);

interface ToastContextWrapper {
  children: React.ReactNode;
}

// Success toasts go away sooner; errors stay long enough to be read.
const DURATIONS: Record<ToastType, number> = { success: 3000, danger: 6000 };

export const ToastContextWrapper = ({ children }: ToastContextWrapper) => {
  const [toasts, setToasts] = useState<ToastItem[]>([]);

  const dismiss = useCallback((id: number) => {
    setToasts((current) => current.filter((toast) => toast.id !== id));
  }, []);

  const showToast = useCallback(
    (message: string, type: ToastType = "success") => {
      const id = Date.now() + Math.random();
      setToasts((current) => [...current, { id, message, type }]);
      window.setTimeout(() => dismiss(id), DURATIONS[type]);
    },
    [dismiss],
  );

  const actions = { showToast };

  return (
    <>
      <ToastContext.Provider value={{ data: {}, setters: {}, actions }}>
        {children}
      </ToastContext.Provider>

      {createPortal(
        <ToastViewport>
          {toasts.map((toast) => (
            <Toast key={toast.id} toast={toast} onClose={() => dismiss(toast.id)} />
          ))}
        </ToastViewport>,
        document.body,
      )}
    </>
  );
};
