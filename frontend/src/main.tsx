import React from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import "@fontsource-variable/inter/wght.css";
import "@fontsource/instrument-serif/400.css";
import "@/lib/i18n";
import { applyTheme, readTheme } from "@/lib/theme";
import "./style.css";

applyTheme(readTheme());

const container = document.getElementById("root");
const root = createRoot(container!);

root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
