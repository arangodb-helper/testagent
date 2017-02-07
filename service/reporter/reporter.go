package reporter

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/arangodb/testAgent/service/chaos"
	"github.com/arangodb/testAgent/service/cluster"
	"github.com/arangodb/testAgent/service/test"
	"github.com/juju/errgo"
	logging "github.com/op/go-logging"
	"golang.org/x/sync/errgroup"
)

const (
	maxChaosEvents = 250
)

type Reporter interface {
	ReportFailure(f test.Failure)
	Reports() []FailureReport
}

type FailureReport struct {
	Failure test.Failure
	Path    string
}

type Service interface {
	Cluster() cluster.Cluster
	Tests() []test.TestScript
	ChaosMonkey() chaos.ChaosMonkey
}

// NewReporter creates a new Reporter using given arguments
func NewReporter(reportDir string, log *logging.Logger, service Service) Reporter {
	return &reporter{
		reportDir: reportDir,
		log:       log,
		service:   service,
	}
}

var (
	fileNameFixer = strings.NewReplacer(
		":", "-",
		"\\", "-",
		"/", "-",
	)
)

type reporter struct {
	reportDir      string
	mutex          sync.Mutex
	log            *logging.Logger
	service        Service
	lastReportID   int32
	failureReports []FailureReport
}

func (s *reporter) Reports() []FailureReport {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return append([]FailureReport{}, s.failureReports...)
}

// ReportFailure report the given failure
func (s *reporter) ReportFailure(f test.Failure) {
	s.log.Infof("Creating failure report for %v", f)
	machines, err := s.service.Cluster().Machines()
	if err != nil {
		s.log.Fatalf("Failed to gather cluster machines: %#v", err)
	}

	// Prepare tmp folder
	folder, err := ioutil.TempDir("", "failure")
	if err != nil {
		s.log.Fatalf("Failed to create temporary failure folder: %#v", err)
	}
	defer os.RemoveAll(folder)
	fileNames := make(chan string)

	// Collect files into tar
	os.MkdirAll(s.reportDir, 0755)
	reportPath := filepath.Join(s.reportDir, s.nextReportID()+".tar")
	tarGroup := errgroup.Group{}
	tarGroup.Go(func() error {
		tarFile, err := os.Create(reportPath)
		if err != nil {
			s.log.Fatalf("Failed to create report file %s: %#v", reportPath, err)
		}
		defer tarFile.Close()
		tw := tar.NewWriter(tarFile)
		for fileName := range fileNames {
			if err := addToTar(s.log, tw, fileName); err != nil {
				s.log.Fatalf("Failed to add %s: %#v", fileName, err)
			}
			tw.Flush()
		}
		tw.Close()
		return nil
	})

	// Generate report file
	if err := s.createFailureReportFile(folder, fileNames, f); err != nil {
		s.log.Fatalf("Failed to create failure file: %#v", err)
	}

	// Collect logs
	if err := s.collectServerLogs(folder, fileNames, machines); err != nil {
		s.log.Fatalf("Failed to collect server logs: %#v", err)
	}

	// Collect cluster state
	if err := s.createClusterStateFile(folder, fileNames, machines); err != nil {
		s.log.Fatalf("Failed to create cluster state: %#v", err)
	}

	// Collect recent chaos
	if err := s.createRecentChaosFile(folder, fileNames); err != nil {
		s.log.Fatalf("Failed to create chaos monkey log: %#v", err)
	}

	// Collect test logs
	if err := s.collectTestLogs(folder, fileNames, s.service.Tests()); err != nil {
		s.log.Fatalf("Failed to collect test logs: %#v", err)
	}

	// Wrap up report file
	close(fileNames)
	if err := tarGroup.Wait(); err != nil {
		s.log.Fatalf("Failed to close report tar file: %#v", err)
	}
	s.log.Infof("Created failure report in %s", reportPath)

	// Record failure
	s.mutex.Lock()
	s.failureReports = append(s.failureReports, FailureReport{
		Failure: f,
		Path:    reportPath,
	})
	s.mutex.Unlock()

	// Notify about failure
}

func (s *reporter) nextReportID() string {
	var id string
	if c := s.service.Cluster(); c != nil {
		id = c.ID()
	}
	index := atomic.AddInt32(&s.lastReportID, 1)
	return fmt.Sprintf("failure-%s-%05d", id, index)
}

// addToTar adds the contents of the given file to the given tar file.
func addToTar(log *logging.Logger, tw *tar.Writer, fileName string) error {
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		log.Errorf("Failed to open %s: %#v", fileName, err)
		return maskAny(err)
	}
	hdr := &tar.Header{
		Name: filepath.Base(fileName),
		Mode: 0644,
		Size: fileInfo.Size(),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		log.Errorf("Failed to write file header for %s: %#v", fileName, err)
		return maskAny(err)
	}
	rd, err := os.Open(fileName)
	if err != nil {
		log.Errorf("Failed to open %s: %#v", fileName, err)
		return maskAny(err)
	}
	defer rd.Close()
	if _, err := io.Copy(tw, rd); err != nil {
		return maskAny(err)
	}
	return nil
}

// collectServerLogs collects recent logs from all servers and adds their filenames to the given channel.
func (s *reporter) collectServerLogs(folder string, fileNames chan string, machines []cluster.Machine) error {
	g := errgroup.Group{}
	for _, m := range machines {
		m := m // Used in nested func
		filePrefix := fileNameFixer.Replace(m.ID())
		if m.HasAgent() {
			g.Go(func() error {
				// Collect agent logs
				if fileName, err := func() (string, error) {
					f, err := os.Create(filepath.Join(folder, fmt.Sprintf("%s-agent.log", filePrefix)))
					if err != nil {
						return "", maskAny(err)
					}
					defer f.Close()
					if err := m.CollectAgentLogs(f); err != nil {
						fmt.Fprintf(f, "\nError fetching logs: %#v\n", err)
						s.log.Errorf("Error fetching agent logs: %#v", err)
					}
					return f.Name(), nil
				}(); err != nil {
					return maskAny(err)
				} else {
					fileNames <- fileName
				}
				return nil
			})
		}
		g.Go(func() error {
			// Collect dbserver logs
			if fileName, err := func() (string, error) {
				f, err := os.Create(filepath.Join(folder, fmt.Sprintf("%s-dbserver.log", filePrefix)))
				if err != nil {
					return "", maskAny(err)
				}
				defer f.Close()
				if err := m.CollectDBServerLogs(f); err != nil {
					fmt.Fprintf(f, "\nError fetching logs: %#v\n", err)
					s.log.Errorf("Error fetching dbserver logs: %#v", err)
				}
				return f.Name(), nil
			}(); err != nil {
				return maskAny(err)
			} else {
				fileNames <- fileName
			}
			return nil
		})
		g.Go(func() error {
			// Collect coordinator logs
			if fileName, err := func() (string, error) {
				f, err := os.Create(filepath.Join(folder, fmt.Sprintf("%s-coordinator.log", filePrefix)))
				if err != nil {
					return "", maskAny(err)
				}
				defer f.Close()
				if err := m.CollectCoordinatorLogs(f); err != nil {
					fmt.Fprintf(f, "\nError fetching logs: %#v\n", err)
					s.log.Errorf("Error fetching coordinator logs: %#v", err)
				}
				return f.Name(), nil
			}(); err != nil {
				return maskAny(err)
			} else {
				fileNames <- fileName
			}
			return nil
		})
		g.Go(func() error {
			// Collect machine logs
			if fileName, err := func() (string, error) {
				f, err := os.Create(filepath.Join(folder, fmt.Sprintf("%s-machine.log", filePrefix)))
				if err != nil {
					return "", maskAny(err)
				}
				defer f.Close()
				if err := m.CollectMachineLogs(f); err != nil {
					fmt.Fprintf(f, "\nError fetching logs: %#v\n", err)
					s.log.Errorf("Error fetching machine logs: %#v", err)
				}
				return f.Name(), nil
			}(); err != nil {
				return maskAny(err)
			} else {
				fileNames <- fileName
			}
			return nil
		})
		g.Go(func() error {
			// Collect network logs
			if fileName, err := func() (string, error) {
				f, err := os.Create(filepath.Join(folder, fmt.Sprintf("%s-network.log", filePrefix)))
				if err != nil {
					return "", maskAny(err)
				}
				defer f.Close()
				if err := m.CollectNetworkLogs(f); err != nil {
					fmt.Fprintf(f, "\nError fetching logs: %#v\n", err)
					s.log.Errorf("Error fetching network logs: %#v", err)
				}
				return f.Name(), nil
			}(); err != nil {
				return maskAny(err)
			} else {
				fileNames <- fileName
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return maskAny(err)
	}

	return nil
}

// collectTestLogs collects logs from all tests and adds their filenames to the given channel.
func (s *reporter) collectTestLogs(folder string, fileNames chan string, tests []test.TestScript) error {
	g := errgroup.Group{}
	for _, t := range tests {
		t := t // Used in nested func
		fileSuffix := fileNameFixer.Replace(t.Name())
		g.Go(func() error {
			// Collect test logs
			if fileName, err := func() (string, error) {
				f, err := os.Create(filepath.Join(folder, fmt.Sprintf("test-%s.log", fileSuffix)))
				if err != nil {
					return "", maskAny(err)
				}
				defer f.Close()
				if err := t.CollectLogs(f); err != nil {
					fmt.Fprintf(f, "\nError fetching logs: %#v\n", err)
					s.log.Errorf("Error fetching test logs: %#v", err)
				}
				return f.Name(), nil
			}(); err != nil {
				return maskAny(err)
			} else {
				fileNames <- fileName
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return maskAny(err)
	}

	return nil
}

// createClusterStateFile dumps the cluster state in a file and adds in the fileNames
func (s *reporter) createClusterStateFile(folder string, fileNames chan string, machines []cluster.Machine) error {
	lines := []string{
		fmt.Sprintf("Cluster state at %s", time.Now()),
		"",
	}
	for _, m := range machines {
		lines = append(lines,
			fmt.Sprintf("Machine %s (%s)", m.ID(), m.State().String()),
		)
		if m.HasAgent() {
			lines = append(lines,
				fmt.Sprintf("Agent url=%v lastReady=%v", urlStr(m.AgentURL()), m.LastAgentReadyStatus()),
			)
		} else {
			lines = append(lines,
				"This machine has no agent",
			)
		}
		lines = append(lines,
			fmt.Sprintf("DBServer url=%v lastReady=%v", urlStr(m.DBServerURL()), m.LastDBServerReadyStatus()),
			fmt.Sprintf("Coordinator url=%v lastReady=%v", urlStr(m.CoordinatorURL()), m.LastCoordinatorReadyStatus()),
			"",
		)

		lines = append(lines, "Network rules")
		rules, err := m.CollectNetworkRules()
		if err != nil {
			lines = append(lines, err.Error())
		} else {
			lines = append(lines, rules...)
		}
		lines = append(lines, "")
	}
	p := filepath.Join(folder, "cluster-state.txt")
	if err := ioutil.WriteFile(p, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return maskAny(err)
	}
	fileNames <- p
	return nil
}

// createRecentChaosFile dumps the recent chaos events in a file and adds in the fileNames
func (s *reporter) createRecentChaosFile(folder string, fileNames chan string) error {
	lines := []string{
		fmt.Sprintf("Recent chaos events at %s", time.Now()),
		"",
	}
	if cm := s.service.ChaosMonkey(); cm != nil {
		lines = append(lines, fmt.Sprintf("Chaos monkey on=%v", cm.Active()))

		stats := cm.Statistics()
		lines = append(lines, "", "Statistics:")
		for _, st := range stats {
			lines = append(lines, fmt.Sprintf("%s: %d", st.Name, st.Value))
		}

		lines = append(lines, "", "Recent events:")
		events := cm.GetRecentEvents(maxChaosEvents)
		for _, e := range events {
			lines = append(lines,
				e.String(),
			)
		}
	} else {
		lines = append(lines,
			"No chaos monkey found",
		)
	}
	p := filepath.Join(folder, "chaos.txt")
	if err := ioutil.WriteFile(p, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return maskAny(err)
	}
	fileNames <- p
	return nil
}

// createFailureReportFile dumps the given failure report in a file and adds in the fileNames
func (s *reporter) createFailureReportFile(folder string, fileNames chan string, f test.Failure) error {
	lines := []string{
		fmt.Sprintf("Failure report at %s", f.Timestamp),
		"",
		f.Message,
	}
	if len(f.Errors) > 0 {
		lines = append(lines,
			"",
			fmt.Sprintf("Error details (%d errors):", len(f.Errors)),
			"",
		)
		for i, err := range f.Errors {
			lines = append(lines,
				fmt.Sprintf("Error %d", i),
				fmt.Sprintf("Message: %v", err),
				fmt.Sprintf("Trace: %#v", errgo.Details(err)),
				"",
			)
		}
	}
	p := filepath.Join(folder, "failure-report.txt")
	if err := ioutil.WriteFile(p, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return maskAny(err)
	}
	fileNames <- p
	return nil
}

func urlStr(u url.URL) string {
	return u.String()
}
