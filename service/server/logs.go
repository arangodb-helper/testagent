package server

import (
	"bytes"
	"fmt"
	"net/http"

	logging "github.com/op/go-logging"
	macaron "gopkg.in/macaron.v1"
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
			var buf bytes.Buffer
			if m.ID() == machineID {
				var err error
				switch mode {
				case "agent":
					err = m.CollectAgentLogs(&buf)
				case "dbserver":
					err = m.CollectDBServerLogs(&buf)
				case "coordinator":
					err = m.CollectCoordinatorLogs(&buf)
				case "machine":
					err = m.CollectMachineLogs(&buf)
				case "network":
					err = m.CollectNetworkLogs(&buf)
				default:
					showError(ctx, fmt.Errorf("Unknown mode '%s'", mode))
					return
				}
				if err == nil {
					ctx.PlainText(http.StatusOK, buf.Bytes())
				} else {
					showError(ctx, fmt.Errorf("Can't collect logs from machine %s", machineID))
				}
				return
			}
		}
	}
	showError(ctx, fmt.Errorf("Unknown machine ID '%s'", machineID))
}
