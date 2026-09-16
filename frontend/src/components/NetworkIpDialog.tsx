import { useContext, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { PrinterContext } from "../contexts/PrinterContext";
import Dialog, { ActionType } from "./Dialog";

const isValidOctet = (value: string) => {
  const number = Number(value);
  return Number.isInteger(number) && number >= 0 && number <= 255;
};

const extractIP = (text: string) => {
  const trimmed = text.trim();
  const match = trimmed.match(/^(?:https?:\/\/)?((?:\d{1,3}\.){3}\d{1,3})(?::\d+)?(?:\/.*)?$/);
  return match?.[1] ?? null;
};

export default function NetworkIpDialog({ openSignal }: { openSignal: number }) {
  const { t } = useTranslation("printers");
  const printerContext = useContext(PrinterContext);
  const [ipParts, setIpParts] = useState(["", "", "", ""]);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

  const updatePart = (index: number, value: string) => {
    const sanitized = value.replace(/\D/g, "").slice(0, 3);

    setIpParts((parts) => {
      const next = [...parts];
      next[index] = sanitized;

      if (next.every((part) => Boolean(part))) {
        setErrorMessage(null);
      }

      return next;
    });

    if ((sanitized.length === 3 || (sanitized.length > 1 && Number(sanitized) > 25)) && index < 3) {
      inputRefs.current[index + 1]?.focus();
      inputRefs.current[index + 1]?.select();
    }
  };

  const handleKeyDown = (event: React.KeyboardEvent<HTMLInputElement>, index: number,) => {
    if (event.ctrlKey || event.metaKey) {
      return;
    }

    const allowedKeys = ["Backspace", "Delete", "ArrowLeft", "ArrowRight", "Tab",];

    if (!/[0-9]/.test(event.key) && !allowedKeys.includes(event.key)) {
      if (event.key === ".") {
        event.preventDefault();

        if (index < 3) {
          inputRefs.current[index + 1]?.focus();
        }
        return;
      }

      event.preventDefault();
    }

    if (event.key === "Backspace" && !ipParts[index] && index > 0) {
      inputRefs.current[index - 1]?.focus();
    }
  };

  const handlePaste = (event: React.ClipboardEvent) => {
    event.preventDefault();
    
    const pasted = event.clipboardData.getData("text").trim();
    const ip = extractIP(pasted);
    const parts = ip?.split(".");
    if (!parts || !parts.every(isValidOctet)) {
        setErrorMessage(t("dialog.invalidPaste"));
        return;
    }
    setIpParts(parts);
    setErrorMessage(null);
    inputRefs.current[3]?.focus();
  };

  const submit = async () => {
    if (!ipParts.every(isValidOctet)) {
        setErrorMessage(t("dialog.invalid"));
        return false;
    }
  
    const ip = ipParts.join(".");
    const result = await printerContext.actions.addLanPrinter(ip);
    if (!result.status) {
      setErrorMessage(result.message);
      return false;
    }

    return true;
  };

  const cleanup = () => {
    setIpParts(["", "", "", ""]);
    setErrorMessage(null);
  };

  return (
    <Dialog
      title={t("dialog.title")}
      actions={[{ name: "submit", label: t("dialog.submit"), disabled: ipParts.some((part) => !part), onClick: submit, variant: "primary" as ActionType }]}
      onClose={cleanup}
      openSignal={openSignal}
    >
      <p className="mb-6 text-[15px] leading-6 text-muted-foreground">
        {t("dialog.description")}
      </p>
      <div
        className="mb-2 flex items-center justify-center gap-2"
        onPasteCapture={handlePaste}
      >
        {ipParts.map((part, index) => (
          <div key={index} className="flex items-center gap-2">
            <input
              ref={(element) => {
                inputRefs.current[index] = element;
                if (index === 0 && !ipParts.some((part) => part)) {
                  element?.focus();
                }
              }}
              value={part}
              type="text"
              inputMode="numeric"
              maxLength={3}
              className="h-12 w-16 rounded-md border border-neutral-300 bg-background text-center text-base focus:border-neutral-500 focus:outline-none focus:ring-4 focus:ring-neutral-200 dark:border-neutral-600 dark:focus:ring-neutral-700"
              onChange={(event) => updatePart(index, event.target.value)}
              onKeyDown={(event) => handleKeyDown(event, index)}
            />

            {index < 3 && (<span className="font-medium text-muted-foreground">.</span>)}
          </div>
        ))}
      </div>

      {errorMessage && (
        <div
          className="mt-4 rounded-md bg-red-50 px-4 py-3 text-sm text-danger dark:bg-red-900/20"
          role="alert"
        >
          {errorMessage}
        </div>
      )}
    </Dialog>
  );
}
