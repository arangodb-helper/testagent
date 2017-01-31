package server

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/arangodb/testAgent/service/chaos"
	"github.com/arangodb/testAgent/service/cluster"
	"github.com/arangodb/testAgent/service/reporter"
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
	Reports() []reporter.FailureReport
}

// StartHTTPServer starts an HTTP server listening on the given port
func StartHTTPServer(log *logging.Logger, port int, reportDir string, service Service) {
	m := macaron.New()
	m.Use(macaron.Logger())
	m.Use(macaron.Recovery())
	m.Use(macaron.Static("",
		macaron.StaticOptions{
			SkipLogging: false,
			FileSystem: bindata.Static(bindata.Options{
				Asset:      templates.Asset,
				AssetDir:   templates.AssetDir,
				AssetInfo:  templates.AssetInfo,
				AssetNames: templates.AssetNames,
				Prefix:     "",
			}),
		},
	))
	m.Use(macaron.Static(reportDir,
		macaron.StaticOptions{
			Prefix:      "reports",
			SkipLogging: false,
		},
	))
	m.Use(macaron.Renderer(macaron.RenderOptions{
		Funcs: []template.FuncMap{
			template.FuncMap{
				"cssReady":  cssReady,
				"cssTestOK": cssTestOK,
			},
		},
		TemplateFileSystem: bindata.Templates(bindata.Options{
			Asset:      templates.Asset,
			AssetDir:   templates.AssetDir,
			AssetInfo:  templates.AssetInfo,
			AssetNames: templates.AssetNames,
			Prefix:     "",
		}),
	}))
	m.Map(log)
	m.Map(service)

	m.Get("/", indexPage)
	m.Get("/test/:name", testPage)
	m.Get("/logs/:machine/:mode", logsPage)

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
