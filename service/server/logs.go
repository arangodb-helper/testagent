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
				switch mode {
				case "agent":
					m.CollectAgentLogs(&buf)
				case "dbserver":
					m.CollectDBServerLogs(&buf)
				case "coordinator":
					m.CollectCoordinatorLogs(&buf)
				case "machine":
					m.CollectMachineLogs(&buf)
				default:
					showError(ctx, fmt.Errorf("Unknown mode '%s'", mode))
					return
				}
				ctx.PlainText(http.StatusOK, buf.Bytes())
				return
			}
		}
	}
	showError(ctx, fmt.Errorf("Unknown machine ID '%s'", machineID))
}
