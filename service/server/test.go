package server

import (
	"bytes"
	"fmt"
	"net/http"

	logging "github.com/op/go-logging"
	macaron "gopkg.in/macaron.v1"
)

func testPage(ctx *macaron.Context, log *logging.Logger, service Service) {
	testName := ctx.Params("name")

	ctests := service.Tests()
	var test Test
	found := false
	for _, ct := range ctests {
		if ct.Name() == testName {
			test = testFromTestScript(ct)
			found = true
			break
		}
	}
	if !found {
		showError(ctx, fmt.Errorf("Test '%s' not found", testName))
	} else {
		ctx.Data["Test"] = test
		ctx.HTML(http.StatusOK, "test")
	}
}

func testLogs(ctx *macaron.Context, log *logging.Logger, service Service) {
	testName := ctx.Params("name")

	ctests := service.Tests()
	var buf bytes.Buffer
	for _, ct := range ctests {
		if ct.Name() == testName {
			if err := ct.CollectLogs(&buf); err != nil {
				showError(ctx, fmt.Errorf("Failed to collect logs from test '%s': %#v", testName, err))
			} else {
				ctx.PlainText(200, buf.Bytes())
				return
			}
		}
	}
	showError(ctx, fmt.Errorf("Test '%s' not found", testName))
}
