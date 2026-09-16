import { main } from "../../wailsjs/go/models";
import { TestPrint } from "../../wailsjs/go/main/App";

// The server prints the test, not the browser: it works even when the interface is opened
// from outside the printer network.
export async function executePrint(
  printer: main.Printer,
  openCashDrawer = false,
) {
  await TestPrint(printer.id, openCashDrawer ? "cashbox" : printer.type);
}
