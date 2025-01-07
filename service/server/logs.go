package server

import (
	"fmt"
	"github.com/op/go-logging"
	"gopkg.in/macaron.v1"
	"net/http"
)

func logsPage(ctx *macaron.Context, log *logging.Logger, service Service) {
	machineID := ctx.Params("machine")
	mode := ctx.Params("mode")

	if c := service.Cluster(); c != nil {
		machines, err := c.Machines()
		if err != nil {
			showError(ctx, err)
			return
		}
		for _, m := range machines {
			if m.ID() == machineID {
				var err error
				ctx.Status(http.StatusOK)
				switch mode {
				case "agent":
					err = m.CollectAgentLogs(ctx.Resp)
				case "dbserver":
					err = m.CollectDBServerLogs(ctx.Resp)
				case "coordinator":
					err = m.CollectCoordinatorLogs(ctx.Resp)
				case "machine":
					err = m.CollectMachineLogs(ctx.Resp)
				case "network":
					err = m.CollectNetworkLogs(ctx.Resp)
				default:
					showError(ctx, fmt.Errorf("Unknown mode '%s'", mode))
					return
				}
				if err != nil {
					showError(ctx, fmt.Errorf("Can't collect logs from machine %s", machineID))
				}
				return
			}
		}
	}
	showError(ctx, fmt.Errorf("Unknown machine ID '%s'", machineID))
}
