import { api } from "../../api";
import { main } from "../models";

export function AppVariable(): Promise<main.AppVariable> {
  return api<main.AppVariable>("/variable");
}

export function Printers(): Promise<main.Printers> {
  return api<main.Printers>("/printers");
}

export function GetTroubleshootInfo(): Promise<main.TroubleshootInfo> {
  return api<main.TroubleshootInfo>("/troubleshoot");
}

export async function IsNetworkPrintingEnabled(): Promise<boolean> {
  const { enabled } = await api<{ enabled: boolean }>("/network-printing");
  return enabled;
}

export async function SetNetworkPrintingEnabled(enabled: boolean): Promise<void> {
  await api("/network-printing", { method: "POST", body: JSON.stringify({ enabled }) });
}

export async function AddLANPrinter(ip: string): Promise<void> {
  await api("/lan-printers", { method: "POST", body: JSON.stringify({ ip }) });
}

// The browser asks for confirmation before removing a printer.
export async function ConfirmRemoveLANPrinter(ip: string): Promise<boolean> {
  if (!window.confirm(`Remove the printer ${ip}?`)) {
    return false;
  }
  await api(`/lan-printers/${encodeURIComponent(ip)}`, { method: "DELETE" });
  return true;
}

export async function CheckLANPrinterStatus(ip: string): Promise<boolean> {
  const { online } = await api<{ online: boolean }>(`/lan-printers/${encodeURIComponent(ip)}/status`);
  return online;
}

export async function TestPrint(id: string, kind: string): Promise<void> {
  await api(`/printers/${encodeURIComponent(id)}/test`, {
    method: "POST",
    body: JSON.stringify({ kind }),
  });
}

export function OdooStatus(): Promise<main.OdooStatus> {
  return api<main.OdooStatus>("/odoo");
}

export function PairWithOdoo(pairing: string, name: string, mode: string): Promise<main.OdooStatus> {
  return api<main.OdooStatus>("/odoo/pair", {
    method: "POST",
    body: JSON.stringify({ pairing, name, mode }),
  });
}

export function UnpairFromOdoo(): Promise<main.OdooStatus> {
  return api<main.OdooStatus>("/odoo/unpair", { method: "POST" });
}
