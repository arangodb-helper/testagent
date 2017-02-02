package server

import (
	"net/http"

	logging "github.com/op/go-logging"
	macaron "gopkg.in/macaron.v1"
)

func chaosPausePage(ctx *macaron.Context, log *logging.Logger, service Service) {
	if cm := service.ChaosMonkey(); cm != nil {
		cm.Stop()
	}
	ctx.Redirect("/", http.StatusFound)
}

func chaosResumePage(ctx *macaron.Context, log *logging.Logger, service Service) {
	if cm := service.ChaosMonkey(); cm != nil {
		cm.Start()
	}
	ctx.Redirect("/", http.StatusFound)
}

func chaosPage(ctx *macaron.Context, log *logging.Logger, service Service) {
	// Chaos
	chaos := chaosFromCluster(service.ChaosMonkey(), 1000)
	log.Debugf("Showing %d chaos events", len(chaos.Events))
	ctx.Data["Chaos"] = chaos

	ctx.HTML(http.StatusOK, "chaos")
}
