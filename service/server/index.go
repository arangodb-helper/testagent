package server

import (
	logging "github.com/op/go-logging"
	macaron "gopkg.in/macaron.v1"
)

type Machine struct {
	CoordinatorURL string
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
	log.Infof("Found %d machines", len(machines))
	ctx.Data["Machines"] = machines
	ctx.HTML(200, "index") // 200 is the response code.
}
