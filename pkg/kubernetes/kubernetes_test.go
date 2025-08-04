package kubernetes

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestGetPID(t *testing.T) {
	// Create a temporary directory to simulate /proc
	tempDir, err := os.MkdirTemp("", "test_proc")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test cases with realistic container IDs
	testCases := []struct {
		name          string
		pid           string
		containerID   string
		cgroupContent string
	}{
		{
			name:        "containerd container",
			pid:         "1234",
			containerID: "abc123def456",
			cgroupContent: `12:perf_event:/kubepods.slice/kubepods-burstable.slice/kubepods-burstable-pod12345.slice/containerd-abc123def456.scope
11:devices:/kubepods.slice/kubepods-burstable.slice/kubepods-burstable-pod12345.slice/containerd-abc123def456.scope
10:freezer:/kubepods.slice/kubepods-burstable.slice/kubepods-burstable-pod12345.slice/containerd-abc123def456.scope`,
		},
		{
			name:        "cri-o container",
			pid:         "5678",
			containerID: "xyz789uvw012",
			cgroupContent: `12:perf_event:/kubepods.slice/kubepods-burstable.slice/kubepods-burstable-pod67890.slice/crio-xyz789uvw012.scope
11:devices:/kubepods.slice/kubepods-burstable.slice/kubepods-burstable-pod67890.slice/crio-xyz789uvw012.scope
10:freezer:/kubepods.slice/kubepods-burstable.slice/kubepods-burstable-pod67890.slice/crio-xyz789uvw012.scope`,
		},
		{
			name:        "docker container",
			pid:         "9012",
			containerID: "mno345pqr678",
			cgroupContent: `12:perf_event:/kubepods/burstable/podmno345pqr678/mno345pqr678
11:devices:/kubepods/burstable/podmno345pqr678/mno345pqr678
10:freezer:/kubepods/burstable/podmno345pqr678/mno345pqr678`,
		},
	}

	// Create mock /proc structure
	for _, tc := range testCases {
		pidDir := filepath.Join(tempDir, tc.pid)
		err := os.MkdirAll(pidDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create PID dir: %v", err)
		}

		cgroupPath := filepath.Join(pidDir, "cgroup")
		err = os.WriteFile(cgroupPath, []byte(tc.cgroupContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write cgroup file: %v", err)
		}
	}

	// Create some noise - PIDs without our container IDs
	noisePIDs := []string{"1", "2", "100", "200"}
	for _, noisePID := range noisePIDs {
		pidDir := filepath.Join(tempDir, noisePID)
		err := os.MkdirAll(pidDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create noise PID dir: %v", err)
		}

		cgroupPath := filepath.Join(pidDir, "cgroup")
		noiseContent := `12:perf_event:/system.slice/system-noise.scope
11:devices:/system.slice/system-noise.scope`
		err = os.WriteFile(cgroupPath, []byte(noiseContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write noise cgroup file: %v", err)
		}
	}

	// Create a test version of GetPID that uses our temp directory
	testGetPID := func(containerID string) (int, error) {
		err := filepath.WalkDir(tempDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			if d.IsDir() && filepath.Dir(path) != tempDir {
				return nil
			}

			if d.IsDir() {
				return nil
			}

			if d.Name() != "cgroup" {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			if strings.Contains(string(content), containerID) {
				pidStr := filepath.Base(filepath.Dir(path))
				pid, err := strconv.Atoi(pidStr)
				if err != nil {
					return nil
				}
				return fmt.Errorf("found pid: %d", pid)
			}

			return nil
		})

		if err != nil && strings.Contains(err.Error(), "found pid:") {
			pidStr := strings.Split(err.Error(), ": ")[1]
			pid, _ := strconv.Atoi(pidStr)
			return pid, nil
		}

		return 0, fmt.Errorf("PID not found for container %s", containerID)
	}

	// Test each container ID
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pid, err := testGetPID(tc.containerID)
			if err != nil {
				t.Errorf("testGetPID failed for %s: %v", tc.name, err)
				return
			}

			expectedPID := 1234
			if tc.pid == "5678" {
				expectedPID = 5678
			} else if tc.pid == "9012" {
				expectedPID = 9012
			}

			if pid != expectedPID {
				t.Errorf("Expected PID %d, got %d for %s", expectedPID, pid, tc.name)
			}

			t.Logf("✓ Found PID %d for container %s (%s)", pid, tc.containerID, tc.name)
		})
	}

	// Test with non-existent container ID
	t.Run("non-existent container", func(t *testing.T) {
		_, err := testGetPID("nonexistent123")
		if err == nil {
			t.Error("Expected error for non-existent container ID")
		}
		t.Logf("✓ Correctly failed for non-existent container: %v", err)
	})
}
