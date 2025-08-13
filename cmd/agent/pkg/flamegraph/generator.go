package flamegraph

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ThomasCardin/gobservability/shared/types"
)

type Generator struct {
	devMode string
}

func NewGenerator(devMode string) *Generator {
	return &Generator{
		devMode: devMode,
	}
}

// GenerateFlamegraph generates a flamegraph for a specific PID
func (g *Generator) GenerateFlamegraph(nodeName, podName string, duration int32, pid int) ([]byte, error) {
	if isDev := os.Getenv(g.devMode); isDev == "true" {
		return []byte(fmt.Sprintf("Mock flamegraph data for node:%s pod:%s duration:%ds",
			nodeName, podName, duration)), nil
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

	// Only generate JSON format
	return g.generateJSONOutput(perfDataFile)
}

// recordPerfData records performance data using perf
func (g *Generator) recordPerfData(outputFile string, duration int32, pid int) error {
	if pid <= 0 {
		return fmt.Errorf("error: invalid PID %d", pid)
	}

	// Check if process exists - try both /proc and /host/proc
	procPath := fmt.Sprintf("/proc/%d", pid)
	hostProcPath := fmt.Sprintf("/host/proc/%d", pid)

	if _, err := os.Stat(procPath); os.IsNotExist(err) {
		if _, err2 := os.Stat(hostProcPath); os.IsNotExist(err2) {
			return fmt.Errorf("error: process with PID %d does not exist in /proc or /host/proc", pid)
		}
		fmt.Printf("process found in /host/proc/%d\n", pid)
	}

	// Test if we can profile this process first
	fmt.Printf("[FLAMEGRAPH] Testing if PID %d can be profiled...\n", pid)
	testCmd := exec.Command("perf", "stat", "-p", strconv.Itoa(pid), "sleep", "1")
	testOutput, testErr := testCmd.CombinedOutput()
	if testErr != nil {
		return fmt.Errorf("error: process PID %d cannot be profiled: %v, output: %s", pid, testErr, string(testOutput))
	}
	fmt.Printf("[FLAMEGRAPH] PID %d can be profiled successfully\n", pid)

	// Use very simple perf command that should work in containers
	cmd := exec.Command("perf", "record", "-F", "99", "-p", strconv.Itoa(pid),
		"-g", "-o", outputFile, "sleep", strconv.Itoa(int(duration)))

	fmt.Printf("[FLAMEGRAPH] running perf command: perf record -F 99 -p %d -g -o %s sleep %d\n", pid, outputFile, duration)

	// Run with a longer timeout to allow perf to finish naturally
	// Use a channel to get the result or timeout
	done := make(chan error, 1)

	go func() {
		// Start the command and capture output
		output, err := cmd.CombinedOutput()
		if err != nil {
			done <- fmt.Errorf("[FLAMEGRAPH] error: perf record failed: %v, output: %s", err, string(output))
		} else {
			fmt.Printf("[FLAMEGRAPH] error: perf record output: %s\n", string(output))
			done <- nil
		}
	}()

	// Wait for completion or timeout
	select {
	case err := <-done:
		if err != nil {
			return err
		}
	case <-time.After(time.Duration(duration*3+60) * time.Second):
		// Kill the process if it's still running (only after a very long time)
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return fmt.Errorf("[FLAMEGRAPH] error: perf record timed out after %d seconds", duration*3+60)
	}

	// Check if perf.data was created
	info, err := os.Stat(outputFile)
	if err != nil || info.Size() == 0 {
		return fmt.Errorf("[FLAMEGRAPH] error: perf.data file was not created or is empty")
	}

	fmt.Printf("[FLAMEGRAPH] perf record completed successfully, file size: %d bytes\n", info.Size())
	return nil
}

// generateFoldedOutput returns the folded stack format
func (g *Generator) generateFoldedOutput(perfDataFile string) ([]byte, error) {
	// Use perf script to get stack traces
	cmd1 := exec.Command("perf", "script", "-i", perfDataFile)
	scriptOutput, err := cmd1.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("[FLAMEGRAPH] error: perf script failed: %v, output: %s", err, string(scriptOutput))
	}

	if len(scriptOutput) == 0 {
		return nil, fmt.Errorf("[FLAMEGRAPH] perf script produced no output")
	}

	// Process with stackcollapse-perf.pl
	cmd2 := exec.Command("stackcollapse-perf.pl")
	cmd2.Stdin = strings.NewReader(string(scriptOutput))
	foldedOutput, err := cmd2.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("[FLAMEGRAPH] error: stackcollapse-perf.pl failed: %v, output: %s", err, string(foldedOutput))
	}

	if len(foldedOutput) == 0 {
		return nil, fmt.Errorf("[FLAMEGRAPH] stackcollapse-perf.pl produced no output")
	}

	return foldedOutput, nil
}

// FlameNode represents a node in the flamegraph JSON structure
type FlameNode struct {
	Name     string      `json:"name"`
	Value    int         `json:"value"`
	Children []FlameNode `json:"children,omitempty"`
}

// generateJSONOutput converts perf data to d3-flame-graph compatible JSON
func (g *Generator) generateJSONOutput(perfDataFile string) ([]byte, error) {
	// First get the folded stack trace data
	foldedData, err := g.generateFoldedOutput(perfDataFile)
	if err != nil {
		return nil, fmt.Errorf("[FLAMEGRAPH] error: failed to generate folded data: %v", err)
	}

	// Parse folded format into hierarchical structure
	root := FlameNode{
		Name:     "root",
		Value:    0,
		Children: []FlameNode{},
	}

	lines := strings.Split(string(foldedData), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse folded format: "func1;func2;func3 count"
		parts := strings.Split(line, " ")
		if len(parts) < 2 {
			continue
		}

		stackStr := parts[0]
		countStr := parts[len(parts)-1]
		count, err := strconv.Atoi(countStr)
		if err != nil {
			continue
		}

		// Split stack trace
		stack := strings.Split(stackStr, ";")
		if len(stack) == 0 {
			continue
		}

		// Add to tree
		g.addStackToTree(&root, stack, count)
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("[FLAMEGRAPH] error: failed to marshal JSON: %v", err)
	}

	return jsonData, nil
}

// addStackToTree adds a stack trace to the flame tree
func (g *Generator) addStackToTree(node *FlameNode, stack []string, count int) {
	if len(stack) == 0 {
		return
	}

	node.Value += count
	funcName := stack[0]
	remaining := stack[1:]

	// Find or create child node
	var child *FlameNode
	for i := range node.Children {
		if node.Children[i].Name == funcName {
			child = &node.Children[i]
			break
		}
	}

	if child == nil {
		// Create new child
		newChild := FlameNode{
			Name:     funcName,
			Value:    0,
			Children: []FlameNode{},
		}
		node.Children = append(node.Children, newChild)
		child = &node.Children[len(node.Children)-1]
	}

	// Recursively add remaining stack
	g.addStackToTree(child, remaining, count)
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
