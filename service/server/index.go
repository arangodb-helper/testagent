package server

import (
	logging "github.com/op/go-logging"
	macaron "gopkg.in/macaron.v1"
)

type Machine struct {
	CoordinatorURL string
}

type Test struct {
	Name     string
	Failures int
	Messages []string
}

func index(ctx *macaron.Context, log *logging.Logger, service Service) {
	machines := []Machine{}
	cluster := service.Cluster()
	if cluster != nil {
		cms, err := service.Cluster().Machines()
		if err != nil {
			showError(ctx, err)
			return
		}
		for _, cm := range cms {
			u := cm.CoordinatorURL()
			machines = append(machines, Machine{
				CoordinatorURL: u.String(),
			})
		}
	}
	log.Debugf("Found %d machines", len(machines))
	ctx.Data["Machines"] = machines

	ctests := service.Tests()
	tests := []Test{}
	for _, ct := range ctests {
		status := ct.Status()
		tests = append(tests, Test{
			Name:     ct.Name(),
			Failures: status.Failures,
			Messages: status.Messages,
		})
	}
	log.Debugf("Found %d tests", len(tests))
	ctx.Data["Tests"] = tests

	ctx.HTML(200, "index") // 200 is the response code.
}
