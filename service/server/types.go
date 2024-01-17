package server

import (
	"path"
	"path/filepath"

	"github.com/arangodb-helper/testagent/service/chaos"
	"github.com/arangodb-helper/testagent/service/cluster"
	"github.com/arangodb-helper/testagent/service/reporter"
	"github.com/arangodb-helper/testagent/service/test"
	humanize "github.com/dustin/go-humanize"
)

type Machine struct {
	ID                         string
	CreatedAt                  string
	StartedAt                  string
	AgentURL                   string
	DBServerURL                string
	CoordinatorURL             string
	HasAgent                   bool
	LastAgentReadyStatus       bool
	LastDBServerReadyStatus    bool
	LastCoordinatorReadyStatus bool
}

type Test struct {
	Name     string
	Active   bool
	Pausing  bool
	Failures int
	Actions  int
	Messages []string
	Counters []test.Counter
}

type Chaos struct {
	Active  bool
	State   string
	Level   int
	Events  []chaos.Event
	Actions []ChaosAction
}

type ChaosAction struct {
	ID        string
	Name      string
	Succeeded int
	Failed    int
	Skipped   int
	Enabled   bool
}

type FailureReport struct {
	Time    string
	Message string
	Path    string
	HRef    string
}

const (
	maxChaosEvents = 20
)

func machineFromCluster(cm cluster.Machine) Machine {
	aURL := cm.AgentURL()
	dURL := cm.DBServerURL()
	cURL := cm.CoordinatorURL()
	return Machine{
		ID:                         cm.ID(),
		CreatedAt:                  humanize.Time(cm.CreatedAt()),
		StartedAt:                  humanize.Time(cm.StartedAt()),
		HasAgent:                   cm.HasAgent(),
		AgentURL:                   aURL.String(),
		DBServerURL:                dURL.String(),
		CoordinatorURL:             cURL.String(),
		LastAgentReadyStatus:       cm.LastAgentReadyStatus(),
		LastDBServerReadyStatus:    cm.LastDBServerReadyStatus(),
		LastCoordinatorReadyStatus: cm.LastCoordinatorReadyStatus(),
	}
}

type machineByID []Machine

// Len is the number of elements in the collection.
func (l machineByID) Len() int { return len(l) }

// Less reports whether the element with
// index i should sort before the element with index j.
func (l machineByID) Less(i, j int) bool { return l[i].ID < l[j].ID }

// Swap swaps the elements with indexes i and j.
func (l machineByID) Swap(i, j int) { l[i], l[j] = l[j], l[i] }

func testFromTestScript(ct test.TestScript) Test {
	status := ct.Status()
	return Test{
		Name:     ct.Name(),
		Active:   status.Active,
		Pausing:  status.Pausing,
		Failures: status.Failures,
		Actions:  status.Actions,
		Messages: status.Messages,
		Counters: status.Counters,
	}
}

func failureReportFromReporter(f reporter.FailureReport) FailureReport {
	return FailureReport{
		Time:    f.Failure.Timestamp.Local().Format("2006-01-02 15:04:05"),
		Message: f.Failure.Message,
		Path:    filepath.Base(f.Path),
		HRef:    "/" + path.Join("reports", filepath.Base(f.Path)),
	}
}

func chaosFromCluster(cm chaos.ChaosMonkey, maxEvents int) Chaos {
	chaos := Chaos{State: "new"}
	if cm != nil {
		chaos.Active = cm.Active()
		chaos.State = cm.State()
		chaos.Events = cm.GetRecentEvents(maxEvents)
		chaos.Level = cm.Level()
		for _, a := range cm.Actions() {
			chaos.Actions = append(chaos.Actions, ChaosAction{
				ID:        a.ID(),
				Name:      a.Name(),
				Succeeded: a.Succeeded(),
				Failed:    a.Failed(),
				Skipped:   a.Skipped(),
				Enabled:   a.Enabled(),
			})
		}
	}
	return chaos
}
