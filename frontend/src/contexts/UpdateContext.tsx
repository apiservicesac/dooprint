import { createContext, useState } from "react";
import { main } from "../../wailsjs/go/models";
import { CheckUpdate, InstallUpdate } from "../../wailsjs/go/main/App";

type UpdateContextType = {
  data: {
    release: main.Release | null;
    checking: boolean;
    installing: boolean;
  };
  actions: {
    check: () => Promise<main.Release>;
    install: () => Promise<void>;
  };
};

export const UpdateContext = createContext({} as UpdateContextType);

interface UpdateContextWrapper {
  children: React.ReactNode;
}

/**
 * Holds what the last check found, so the banner and the System tab show the same thing. The
 * device asks GitHub only when check() is called, that is, when someone presses the button.
 */
export const UpdateContextWrapper = ({ children }: UpdateContextWrapper) => {
  const [release, setRelease] = useState<main.Release | null>(null);
  const [checking, setChecking] = useState(false);
  const [installing, setInstalling] = useState(false);

  const check = async () => {
    setChecking(true);
    try {
      const found = await CheckUpdate();
      setRelease(found);
      return found;
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
    <UpdateContext.Provider value={{ data: { release, checking, installing }, actions: { check, install } }}>
      {children}
    </UpdateContext.Provider>
  );
};
