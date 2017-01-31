package server

import (
	"net/http"
	"sort"

	humanize "github.com/dustin/go-humanize"
	logging "github.com/op/go-logging"
	macaron "gopkg.in/macaron.v1"
)

func indexPage(ctx *macaron.Context, log *logging.Logger, service Service) {
	// General
	ctx.Data["Uptime"] = humanize.Time(service.StartedAt())

	// Cluster
	machines := []Machine{}
	cluster := service.Cluster()
	if cluster != nil {
		ctx.Data["ID"] = cluster.ID()
		cms, err := service.Cluster().Machines()
		if err != nil {
			showError(ctx, err)
			return
		}
		for _, cm := range cms {
			machines = append(machines, machineFromCluster(cm))
		}
	}
	sort.Sort(machineByID(machines))
	log.Debugf("Found %d machines", len(machines))
	ctx.Data["Machines"] = machines

	// Tests
	ctests := service.Tests()
	tests := []Test{}
	for _, ct := range ctests {
		tests = append(tests, testFromTestScript(ct))
	}
	log.Debugf("Found %d tests", len(tests))
	ctx.Data["Tests"] = tests

	// Chaos
	cm := service.ChaosMonkey()
	chaos := Chaos{State: "new"}
	if cm != nil {
		chaos.Active = cm.Active()
		chaos.State = cm.State()
		chaos.Events = cm.GetRecentEvents()
		if len(chaos.Events) > maxChaosEvents {
			chaos.Events = chaos.Events[:maxChaosEvents]
		}
	}
	log.Debugf("Found %d chaos events", len(chaos.Events))
	ctx.Data["Chaos"] = chaos

	// Failure reports
	creports := service.Reports()
	reports := []FailureReport{}
	for _, r := range creports {
		reports = append(reports, failureReportFromReporter(r))
	}
	log.Debugf("Found %d failure reports", len(reports))
	ctx.Data["Reports"] = reports

	ctx.HTML(http.StatusOK, "index")
}
