package server

import (
	"fmt"
	"net/http"

	"github.com/arangodb/testAgent/service/chaos"
	"github.com/arangodb/testAgent/service/cluster"
	"github.com/arangodb/testAgent/service/test"
	"github.com/arangodb/testAgent/templates"
	"github.com/go-macaron/bindata"
	logging "github.com/op/go-logging"
	macaron "gopkg.in/macaron.v1"
)

type Service interface {
	Cluster() cluster.Cluster
	Tests() []test.TestScript
	ChaosMonkey() chaos.ChaosMonkey
}

// StartHTTPServer starts an HTTP server listening on the given port
func StartHTTPServer(log *logging.Logger, port int, service Service) {
	m := macaron.Classic()
	m.Use(macaron.Renderer(macaron.RenderOptions{
		TemplateFileSystem: bindata.Templates(bindata.Options{
			Asset:      templates.Asset,
			AssetDir:   templates.AssetDir,
			AssetNames: templates.AssetNames,
			Prefix:     "",
		}),
	}))
	m.Map(log)
	m.Map(service)

	m.Get("/", index)

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	log.Infof("HTTP server listening on %s", addr)
	go func() {
		if err := http.ListenAndServe(addr, m); err != nil {
			log.Fatalf("Failed to start listener: %#v", err)
		}
	}()
}

func showError(ctx *macaron.Context, err error) {
	msg := err.Error()
	ctx.PlainText(http.StatusInternalServerError, []byte(msg))
}
