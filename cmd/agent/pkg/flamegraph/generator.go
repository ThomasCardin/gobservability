package flamegraph

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ThomasCardin/peek/shared/types"
)

type Generator struct {
	devMode string
}

func NewGenerator(devMode string) *Generator {
	return &Generator{
		devMode: devMode,
	}
}

// GenerateFlamegraph generates a flamegraph for a specific PID or entire node
func (g *Generator) GenerateFlamegraph(nodeName, podName string, duration int32, format string, pid int) ([]byte, error) {
	if isDev := os.Getenv(g.devMode); isDev == "true" {
		// Return mock data in dev mode
		return []byte(fmt.Sprintf("Mock flamegraph data for node:%s pod:%s duration:%ds format:%s",
			nodeName, podName, duration, format)), nil
	}

	// Create temporary directory for flamegraph files
	tmpDir, err := os.MkdirTemp("", "flamegraph-*")
	if err != nil {
		return nil, fmt.Errorf("error: failed to create temp directory %s", err.Error())
	}
	defer os.RemoveAll(tmpDir)

	// Generate perf data
	perfDataFile := filepath.Join(tmpDir, "perf.data")
	if err := g.recordPerfData(perfDataFile, duration, pid); err != nil {
		return nil, fmt.Errorf("error: failed to record perf data %s", err.Error())
	}

	// Convert perf data to flamegraph format
	switch strings.ToLower(format) {
	case "svg":
		return g.generateSVGFlamegraph(perfDataFile, tmpDir)
	case "folded":
		return g.generateFoldedOutput(perfDataFile)
	case "txt":
		return g.generateTextOutput(perfDataFile)
	default:
		return nil, fmt.Errorf("error: unsupported flamegraph format: %s", format)
	}
}

// recordPerfData records performance data using perf
func (g *Generator) recordPerfData(outputFile string, duration int32, pid int) error {
	if pid <= 0 {
		return fmt.Errorf("error: invalid PID %d", pid)
	}

	// Record for specific PID only
	cmd := exec.Command("perf", "record", "-F", "99", "-p", strconv.Itoa(pid),
		"-g", "-o", outputFile, "--", "sleep", strconv.Itoa(int(duration)))

	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("perf record failed: %v", err)
	}

	return nil
}

// generateSVGFlamegraph creates an SVG flamegraph
func (g *Generator) generateSVGFlamegraph(perfDataFile, tmpDir string) ([]byte, error) {
	// First convert perf.data to folded format
	foldedFile := filepath.Join(tmpDir, "folded.txt")
	cmd1 := exec.Command("perf", "script", "-i", perfDataFile)
	cmd2 := exec.Command("stackcollapse-perf.pl")
	cmd3 := exec.Command("tee", foldedFile)

	// Chain commands: perf script | stackcollapse-perf.pl | tee folded.txt
	cmd2.Stdin, _ = cmd1.StdoutPipe()
	cmd3.Stdin, _ = cmd2.StdoutPipe()

	if err := cmd1.Start(); err != nil {
		return nil, fmt.Errorf("failed to start perf script: %v", err)
	}
	if err := cmd2.Start(); err != nil {
		return nil, fmt.Errorf("failed to start stackcollapse: %v", err)
	}
	if err := cmd3.Start(); err != nil {
		return nil, fmt.Errorf("failed to start tee: %v", err)
	}

	if err := cmd1.Wait(); err != nil {
		return nil, fmt.Errorf("perf script failed: %v", err)
	}
	if err := cmd2.Wait(); err != nil {
		return nil, fmt.Errorf("stackcollapse failed: %v", err)
	}
	if err := cmd3.Wait(); err != nil {
		return nil, fmt.Errorf("tee failed: %v", err)
	}

	// Generate SVG from folded data
	cmd4 := exec.Command("flamegraph.pl", foldedFile)
	output, err := cmd4.Output()
	if err != nil {
		return nil, fmt.Errorf("flamegraph.pl failed: %v", err)
	}

	return output, nil
}

// generateFoldedOutput returns the folded stack format
func (g *Generator) generateFoldedOutput(perfDataFile string) ([]byte, error) {
	cmd1 := exec.Command("perf", "script", "-i", perfDataFile)
	cmd2 := exec.Command("stackcollapse-perf.pl")

	cmd2.Stdin, _ = cmd1.StdoutPipe()

	if err := cmd1.Start(); err != nil {
		return nil, fmt.Errorf("failed to start perf script: %v", err)
	}
	if err := cmd2.Start(); err != nil {
		return nil, fmt.Errorf("failed to start stackcollapse: %v", err)
	}

	output, err := cmd2.Output()
	if err != nil {
		return nil, fmt.Errorf("stackcollapse failed: %v", err)
	}

	if err := cmd1.Wait(); err != nil {
		return nil, fmt.Errorf("perf script failed: %v", err)
	}

	return output, nil
}

// generateTextOutput returns raw perf script output
func (g *Generator) generateTextOutput(perfDataFile string) ([]byte, error) {
	cmd := exec.Command("perf", "script", "-i", perfDataFile)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("perf script failed: %v", err)
	}

	return output, nil
}

// GetPIDForPod finds the PID for a specific pod name
func (g *Generator) GetPIDForPod(podName string, pods []*types.Pod) int {
	if podName == "" {
		return -1 // No system-wide flamegraph support
	}

	for _, pod := range pods {
		if pod.Name == podName && pod.PID > 0 {
			return pod.PID
		}
	}

	return -1
}
