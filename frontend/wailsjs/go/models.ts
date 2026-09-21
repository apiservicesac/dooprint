// Types matching the Go structures returned by the API (internal/app).
export namespace main {
  export interface Printer {
    name: string;
    ip: string;
    id: string;
    isLAN: boolean;
    lanIp?: string;
    online: boolean;
    type: string;
  }

  export interface UnavailablePrinter {
    name: string;
    errorMsg: string;
    isLAN: boolean;
    lanIp?: string;
  }

  export interface Printers {
    errorMsg: string;
    printers: Printer[];
    unavailablePrinters: UnavailablePrinter[];
  }

  export interface AppVariable {
    serverRunning: boolean;
    os: string;
    mode: string;
    port: number;
    name: string;
    version: string;
    address: string;
  }

  export interface OdooStatus {
    paired: boolean;
    odooUrl: string;
    boxName: string;
    mode: string;
    busState: string;
    lastPoll: string;
    lastError: string;
    printed: number;
    failed: number;
  }

  export interface Release {
    current: string;
    latest: string;
    newer: boolean;
    notes?: string;
  }

  export interface TroubleshootInfo {
    activeFirewall: string;
    firewallZone: string;
    port: number;
    subnet: string;
    localIp: string;
    execPath: string;
  }
}
