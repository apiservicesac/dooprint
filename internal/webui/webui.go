// Package webui serves the configuration web interface and its API on the ePOS server port.
// The React frontend is embedded in the binary and talks to these routes.
package webui

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/apiservicesac/dooprint/internal/app"
	"github.com/apiservicesac/dooprint/internal/escpos"
	"github.com/apiservicesac/dooprint/internal/printer"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

//go:embed all:dist
var assets embed.FS

// Jobs sent by the test button of the interface, one per printer type.
const (
	testReceipt = `<epos-print>
	<text align="center" dw="true" dh="true">TEST&#10;</text>
	<text align="center">%s&#10;</text>
	<feed line="2" />
	<cut type="feed" />
</epos-print>`
	testCashbox = `<epos-print><pulse /></epos-print>`
	testLabel   = "^XA^PW800^LL250^CF0,35^FO0,40^FB800,1,0,C^FDTEST^FS^CF0,25^FO0,100^FB800,1,0,C^FD%s^FS^XZ"
)

func Register(service *app.Service, manager *printer.Manager) func(fiber.Router) {
	return func(router fiber.Router) {
		api := router.Group("/api")

		api.Get("/variable", func(ctx fiber.Ctx) error {
			return ctx.JSON(service.Variable())
		})
		api.Get("/printers", func(ctx fiber.Ctx) error {
			return ctx.JSON(service.Printers())
		})
		api.Get("/odoo", func(ctx fiber.Ctx) error {
			return ctx.JSON(service.LinkStatus())
		})
		api.Post("/odoo/pair", func(ctx fiber.Ctx) error {
			var body struct {
				Pairing string `json:"pairing"`
				Name    string `json:"name"`
				Mode    string `json:"mode"`
			}
			if err := ctx.Bind().Body(&body); err != nil {
				return fiber.NewError(http.StatusBadRequest, err.Error())
			}
			if service.Link == nil {
				return fiber.NewError(http.StatusInternalServerError, "enlace no disponible")
			}
			if err := service.Link.Pair(body.Pairing, body.Name, body.Mode); err != nil {
				return fiber.NewError(http.StatusBadRequest, err.Error())
			}
			return ctx.JSON(service.LinkStatus())
		})
		api.Post("/odoo/unpair", func(ctx fiber.Ctx) error {
			if service.Link != nil {
				if err := service.Link.Unpair(); err != nil {
					return fiber.NewError(http.StatusInternalServerError, err.Error())
				}
			}
			return ctx.JSON(service.LinkStatus())
		})
		api.Get("/troubleshoot", func(ctx fiber.Ctx) error {
			return ctx.JSON(service.Troubleshoot())
		})
		api.Get("/network-printing", func(ctx fiber.Ctx) error {
			return ctx.JSON(fiber.Map{"enabled": service.Config.IsNetworkPrintingEnabled()})
		})
		api.Post("/network-printing", func(ctx fiber.Ctx) error {
			var body struct {
				Enabled bool `json:"enabled"`
			}
			if err := ctx.Bind().Body(&body); err != nil {
				return fiber.NewError(http.StatusBadRequest, err.Error())
			}
			if err := service.Config.SetNetworkPrintingEnabled(body.Enabled); err != nil {
				return fiber.NewError(http.StatusInternalServerError, err.Error())
			}
			return ctx.JSON(fiber.Map{"enabled": body.Enabled})
		})
		api.Post("/lan-printers", func(ctx fiber.Ctx) error {
			var body struct {
				Ip string `json:"ip"`
			}
			if err := ctx.Bind().Body(&body); err != nil {
				return fiber.NewError(http.StatusBadRequest, err.Error())
			}
			if err := service.AddLANPrinter(body.Ip); err != nil {
				return fiber.NewError(http.StatusBadRequest, err.Error())
			}
			return ctx.JSON(fiber.Map{"ip": body.Ip})
		})
		api.Delete("/lan-printers/:ip", func(ctx fiber.Ctx) error {
			if err := service.RemoveLANPrinter(ctx.Params("ip")); err != nil {
				return fiber.NewError(http.StatusInternalServerError, err.Error())
			}
			return ctx.JSON(fiber.Map{"removed": ctx.Params("ip")})
		})
		api.Get("/lan-printers/:ip/status", func(ctx fiber.Ctx) error {
			return ctx.JSON(fiber.Map{"online": service.CheckLANPrinterStatus(ctx.Params("ip"))})
		})
		api.Post("/printers/:id/test", func(ctx fiber.Ctx) error {
			var body struct {
				Kind string `json:"kind"`
			}
			if err := ctx.Bind().Body(&body); err != nil {
				return fiber.NewError(http.StatusBadRequest, err.Error())
			}
			job, err := testJob(body.Kind, service.PrinterURL(ctx.Params("id")))
			if err != nil {
				return fiber.NewError(http.StatusInternalServerError, err.Error())
			}
			reply, err := manager.WriteAsync(ctx.Params("id"), job)
			if err == nil {
				if result := <-reply; !result.OK {
					err = result.Err
				}
			}
			if err != nil {
				return fiber.NewError(http.StatusBadGateway, err.Error())
			}
			return ctx.JSON(fiber.Map{"printed": true})
		})

		dist, err := fs.Sub(assets, "dist")
		if err != nil {
			panic(err)
		}
		router.Use("/", static.New("", static.Config{FS: dist, IndexNames: []string{"index.html"}}))
	}
}

// testJob builds the test job: label printers take raw ZPL, the rest take ePOS, which is
// converted to ESC/POS right here.
func testJob(kind, printerURL string) ([]byte, error) {
	if kind == "label" {
		return []byte(fmt.Sprintf(testLabel, printerURL)), nil
	}
	if kind == "cashbox" {
		return escpos.ParseXML([]byte(testCashbox))
	}
	return escpos.ParseXML([]byte(fmt.Sprintf(testReceipt, printerURL)))
}
