import { createContext, useEffect, useState } from "react";
import { main } from "../../wailsjs/go/models";
import { CheckUpdate, InstallUpdate, UpdateStatus } from "../../wailsjs/go/main/App";

type UpdateContextType = {
  data: {
    status: main.UpdateStatus | null;
    checking: boolean;
    installing: boolean;
  };
  actions: {
    check: () => Promise<main.UpdateStatus | null>;
    install: () => Promise<void>;
  };
};

export const UpdateContext = createContext({} as UpdateContextType);

interface UpdateContextWrapper {
  children: React.ReactNode;
}

/**
 * Keeps what is known about updates. The device only asks GitHub when someone presses the
 * button, so on start this just reads the result of the last check of this run.
 */
export const UpdateContextWrapper = ({ children }: UpdateContextWrapper) => {
  const [status, setStatus] = useState<main.UpdateStatus | null>(null);
  const [checking, setChecking] = useState(false);
  const [installing, setInstalling] = useState(false);

  useEffect(() => {
    UpdateStatus()
      .then(setStatus)
      .catch((error) => console.error("Cannot read the update status", error));
  }, []);

  const check = async () => {
    setChecking(true);
    try {
      const result = await CheckUpdate();
      setStatus(result);
      return result;
    } finally {
      setChecking(false);
    }
  };

  const install = async () => {
    setInstalling(true);
    try {
      await InstallUpdate();
    } finally {
      setInstalling(false);
    }
  };

  return (
    <UpdateContext.Provider value={{ data: { status, checking, installing }, actions: { check, install } }}>
      {children}
    </UpdateContext.Provider>
  );
};
