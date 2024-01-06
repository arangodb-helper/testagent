package metrics

import (
	"fmt"
	"os"

	dc "github.com/fsouza/go-dockerclient"
)

type DockerMetricsWriter interface {
	Write() error
}

type fileMetricsWriter struct {
	stats chan *dc.Stats
	done  chan bool
	file  string
}

func NewDockerMetricsWriter(stats chan *dc.Stats, done chan bool, file string) DockerMetricsWriter {
	return &fileMetricsWriter{
		stats: stats,
		done:  done,
		file:  file,
	}
}

func (w *fileMetricsWriter) Write() error {
	file, err := os.Create(w.file)
	if err != nil {
		return err
	}
	defer file.Close()
	file.WriteString("timestamp,cpu_total_usage,cpu_usage_in_kernelmode,cpu_usage_in_usermode,system_cpu_usage,memory_usage\n")
	for stat := range w.stats {
		if _, err := fmt.Fprintf(file, "%d,%d,%d,%d,%d,%d\n", stat.Read.UnixNano(), stat.CPUStats.CPUUsage.TotalUsage, stat.CPUStats.CPUUsage.UsageInKernelmode, stat.CPUStats.CPUUsage.UsageInUsermode, stat.CPUStats.SystemCPUUsage, stat.MemoryStats.Usage); err != nil {
			return err
		}
	}
	return nil
}
