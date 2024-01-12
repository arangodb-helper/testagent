package metrics

import (
	"errors"
	"fmt"
	"os"

	dc "github.com/fsouza/go-dockerclient"
	logging "github.com/op/go-logging"
)

type DockerMetricsWriter interface {
	Write() error
}

type fileMetricsWriter struct {
	stats chan *dc.Stats
	done  chan bool
	file  string
	log   *logging.Logger
}

func NewDockerMetricsWriter(stats chan *dc.Stats, done chan bool, file string, log *logging.Logger) DockerMetricsWriter {
	return &fileMetricsWriter{
		stats: stats,
		done:  done,
		file:  file,
		log:   log,
	}
}

func (w *fileMetricsWriter) createOrOpenExistingFile() (*os.File, error) {
	fileExists := true
	if _, err := os.Stat(w.file); errors.Is(err, os.ErrNotExist) {
		fileExists = false
	}
	file, err := os.OpenFile(w.file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	if !fileExists {
		file.WriteString("timestamp,cpu_total_usage,cpu_usage_in_kernelmode,cpu_usage_in_usermode,system_cpu_usage,memory_usage\n")
	}
	return file, nil
}

func (w *fileMetricsWriter) Write() error {
	file, err := w.createOrOpenExistingFile()
	if err != nil {
		return err
	}
	defer file.Close()
	w.log.Infof("Starting writing metrics to file: %s", w.file)
	for stat := range w.stats {
		if stat.PidsStats.Current == 0 {
			// if the container is not alive anymore,
			// it probably was killed/restarted by chaos monkey
			// in this case we must stop writing metrics and destroy the writer object
			// writing must be restarted again when the container is back online
			w.done <- true
			w.log.Infof("Stopping writing metrics to file: %s", w.file)
			return nil
		}
		if _, err := fmt.Fprintf(file, "%d,%d,%d,%d,%d,%d\n", stat.Read.UnixNano(), stat.CPUStats.CPUUsage.TotalUsage, stat.CPUStats.CPUUsage.UsageInKernelmode, stat.CPUStats.CPUUsage.UsageInUsermode, stat.CPUStats.SystemCPUUsage, stat.MemoryStats.Usage); err != nil {
			return err
		}
	}
	return nil
}
