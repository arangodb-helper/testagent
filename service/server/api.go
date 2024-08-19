package server

import (
	"net/http"
	"strconv"

	logging "github.com/op/go-logging"
	macaron "gopkg.in/macaron.v1"
)

func failureCount(ctx *macaron.Context, log *logging.Logger, service Service) {
	failureReports := service.Reports()
	failureReportsNumber := len(failureReports)
	ctx.PlainText(http.StatusOK, []byte(strconv.Itoa(failureReportsNumber)))
}
