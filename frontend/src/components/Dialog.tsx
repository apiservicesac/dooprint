import { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { useMountTransition } from "../hooks/useMountTransition";
import CloseButton from "./CloseButton";

export type ActionType = "primary" | "secondary" | "danger"

export interface DialogAction {
  name: string;
  label: string;
  onClick?: (helpers: { close: () => void }) => void | boolean | Promise<boolean | void>;
  disabled?: boolean;
  variant?: ActionType;
  className?: string;
}

interface DialogProps {
  title: string;
  openButton?: React.ReactNode;
  children: React.ReactNode;
  actions?: DialogAction[];
  onClose?: () => void;
  onOpen?: () => void;
  showTitleDivider?: boolean;
  /** Bump this value (e.g. a counter) to open the dialog programmatically, without an openButton. */
  openSignal?: number;
}

export default function Dialog({
  title,
  children,
  openButton,
  actions = [],
  onClose,
  onOpen,
  showTitleDivider = false,
  openSignal,
}: DialogProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [loadingAction, setLoadingAction] = useState<string | null>(null);
  const { mounted } = useMountTransition(isOpen);
  const previousOpenSignal = useRef(openSignal);

  const close = () => {
    setIsOpen(false);
    onClose?.();
  };

  const open = () => {
    setIsOpen(true);
    onOpen?.();
  };

  useEffect(() => {
    if (openSignal === undefined || openSignal === previousOpenSignal.current) {
      return;
    }

    previousOpenSignal.current = openSignal;
    open();
  }, [openSignal]);

  const isExecuting = useRef(false);

  const handleAction = async (action: DialogAction) => {
    if (action.disabled || isExecuting.current) {
      return;
    }

    if (!action.onClick) {
      close();
      return;
    }

    isExecuting.current = true;
    setLoadingAction(action.name);
    try {
      const result = await action.onClick({ close });
      if (result !== false) {
        close();
      }
    } finally {
      isExecuting.current = false;
      setLoadingAction(null);
    }
  };

  useEffect(() => {
    if (!isOpen) {
      return;
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        close();
        return;
      }

      if (event.key !== "Enter" || event.repeat) {
        return;
      }

      if ((event.target as HTMLElement | null)?.closest("button, a")) {
        return;
      }

      const primaryAction = actions.find((action) => !action.disabled && action.variant !== "secondary");
      if (primaryAction && !isExecuting.current) {
        handleAction(primaryAction);
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => {
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, [isOpen, actions, onClose]);

  return (
    <>
      {openButton && <div onClick={() => open()}>{openButton}</div>}
      {mounted &&
        createPortal(
          <div
            className={`fixed inset-0 z-50 flex items-end sm:items-center justify-center p-4 transition ${isOpen
              ? "opacity-100 duration-200 ease-out"
              : "opacity-0 duration-150 ease-in"
              }`}
          >
            <div
              className="absolute inset-0 bg-black/40 backdrop-blur-[2px]"
              onClick={() => close()}
            />

            <div className="relative w-full max-w-md max-h-[calc(100vh-2rem)] overflow-y-auto overflow-x-hidden rounded-lg border border-border bg-card p-7 text-card-foreground shadow-xl">
              <div className={`flex items-center justify-between ${showTitleDivider ? "mb-5 border-b border-border pb-4" : "mb-5"}`}>
                <div className="text-lg font-semibold">{title}</div>
                <CloseButton onClick={() => close()} />
              </div>

              <div>{children}</div>

              {actions.length > 0 && (
                <div className="flex items-center gap-3 pt-6">
                  {actions.map((action) => {
                    const isActionLoading = loadingAction === action.name;
                    const defaultVariantClass =
                      action.variant === "secondary"
                        ? "btn-secondary"
                        : action.variant === "danger"
                          ? "btn-red"
                          : "btn-primary";

                    return (
                      <button
                        key={action.name}
                        type="button"
                        disabled={Boolean(action.disabled) || isActionLoading || Boolean(loadingAction)}
                        className={action.className ?? `${defaultVariantClass} flex-1 sm:w-full`}
                        onClick={() => handleAction(action)}
                      >
                        {isActionLoading ? "…" : action.label}
                      </button>
                    );
                  })}
                </div>
              )}
            </div>
          </div>,
          document.body,
        )}
    </>
  );
}
