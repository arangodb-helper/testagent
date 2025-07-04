package server

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/arangodb-helper/testagent/service/chaos"
	"github.com/arangodb-helper/testagent/service/cluster"
	"github.com/arangodb-helper/testagent/service/reporter"
	"github.com/arangodb-helper/testagent/service/test"
	"github.com/arangodb-helper/testagent/templates"
	"github.com/go-macaron/bindata"
	logging "github.com/op/go-logging"
	macaron "gopkg.in/macaron.v1"
)

type Service interface {
	StartedAt() time.Time
	ProjectVersion() string
	ProjectBuild() string
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
				"cssReady":   cssReady,
				"cssTestOK":  cssTestOK,
				"formatTime": formatTime,
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

	// UI
	m.Get("/", indexPage)
	m.Get("/test/:name", testPage)
	m.Get("/test/:name/pause", testPausePage)
	m.Get("/test/:name/resume", testResumePage)
	m.Get("/test/:name/logs", testLogs)
	m.Get("/logs/:machine/:mode", logsPage)
	m.Get("/chaos", chaosPage)
	m.Get("/chaos/pause", chaosPausePage)
	m.Get("/chaos/resume", chaosResumePage)
	m.Get("/chaos/:id/enable", chaosActionEnablePage)
	m.Get("/chaos/:id/disable", chaosActionDisablePage)
	m.Get("/chaos/level/:level", chaosSetLevel)
	m.Get("/api/reportMessage/:idx", reportMessage)

	// API
	m.Get("/api/failureCount", failureCount)
	m.Post("/api/pauseAllTests", pauseAllTests)
	m.Post("/api/resumeAllTests", resumeAllTests)

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
