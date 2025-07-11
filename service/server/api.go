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

func reportMessage(ctx *macaron.Context, log *logging.Logger, service Service) {
	failureReports := service.Reports()
	reportIdx, err := strconv.Atoi(ctx.Params("idx"))
	if err != nil {
		ctx.PlainText(http.StatusBadRequest, []byte("invalid report ID"))
		return
	}
	if reportIdx < 0 || reportIdx >= len(failureReports) {
		ctx.PlainText(http.StatusNotFound, []byte("report not found"))
		return
	}
	failureReport := failureReports[reportIdx]
	ctx.PlainText(http.StatusOK, []byte(failureReport.Failure.Message))
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
		ctx.PlainText(http.StatusInternalServerError, []byte("error while pausing tests! see server log for details."))
	} else {
		ctx.PlainText(http.StatusOK, []byte("OK"))
	}
}

func resumeAllTests(ctx *macaron.Context, log *logging.Logger, service Service) {
	tests := service.Tests()
	failed := false
	for _, test := range tests {
		err := test.Resume()
		if err != nil {
			failed = true
		}
	}
	if failed {
		ctx.PlainText(http.StatusInternalServerError, []byte("error while resuming tests! see server log for details."))
	} else {
		ctx.PlainText(http.StatusOK, []byte("OK"))
	}
}
