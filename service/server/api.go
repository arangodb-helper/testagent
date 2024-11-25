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

func pauseAllTests(ctx *macaron.Context, log *logging.Logger, service Service) {
	tests := service.Tests()
	failed := false
	for _, test := range tests {
		err := test.Pause()
		if err != nil {
			failed = true
		}
	}
	if failed {
		ctx.PlainText(http.StatusInternalServerError, []byte("error! see server log for details."))
	} else {
		ctx.PlainText(http.StatusOK, []byte("OK"))
	}
}
