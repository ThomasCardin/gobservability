package kubernetes

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ThomasCardin/gobservability/cmd/agent/internal"
	"github.com/ThomasCardin/gobservability/cmd/agent/shared"
	"github.com/ThomasCardin/gobservability/shared/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func GetPodsPID(devMode, nodeName string) ([]*types.Pod, error) {
	if isDev := os.Getenv(devMode); isDev == "true" {
		return generateFakePods(nodeName), nil
	}

	// In-cluster config for production
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName,
	})
	if err != nil {
		return nil, err
	}

	var result []*types.Pod
	for _, pod := range pods.Items {
		// Get containerid
		containerIDs, err := getContainerID(pod.Name, pod.Status.ContainerStatuses)
		if err != nil {
			result = append(result, &types.Pod{
				Name:        pod.Name,
				ContainerID: "Not found",
				PID:         -1,
			})
			continue
		}

		// Get PID and collect metrics
		for _, containerID := range containerIDs {
			pid, err := getPID(devMode, containerID)
			if err != nil {
				result = append(result, &types.Pod{
					Name:        pod.Name,
					ContainerID: containerID,
					PID:         -1,
					PodMetrics:  types.PodMetrics{}, // Empty metrics for failed pods
					PidDetails:  types.PidDetails{}, // Empty details for failed pods
				})
			} else {
				// Collect detailed metrics for this PID
				podMetrics, pidDetails, metricsErr := internal.CollectPodMetrics(devMode, pid)
				if metricsErr != nil {
					result = append(result, &types.Pod{
						Name:        pod.Name,
						ContainerID: containerID,
						PID:         pid,
						PodMetrics:  types.PodMetrics{},
						PidDetails:  types.PidDetails{},
					})
				} else {
					result = append(result, &types.Pod{
						Name:        pod.Name,
						ContainerID: containerID,
						PID:         pid,
						PodMetrics:  *podMetrics,
						PidDetails:  *pidDetails,
					})
				}
			}
		}
	}

	return result, nil
}

func getContainerID(podName string, containerStatuses []v1.ContainerStatus) ([]string, error) {
	var containers []string

	// Get container IDs from status
	for _, containerStatus := range containerStatuses {
		if containerStatus.ContainerID != "" {
			// Remove runtime prefix (docker://, containerd://, etc.)
			containerID := containerStatus.ContainerID
			if idx := strings.LastIndex(containerID, "://"); idx != -1 {
				containerID = containerID[idx+3:]
			}
			containers = append(containers, containerID)
		}
	}

	if len(containers) == 0 {
		return nil, fmt.Errorf("no containers found for pod %s", podName)
	}

	return containers, nil
}

func getPID(devMode, containerID string) (int, error) {
	// Parse /proc/*/cgroup to find PID from container ID
	procBasePath := shared.GetProcBasePath(devMode)
	file, err := os.Open(procBasePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	dirs, err := file.Readdir(-1)
	if err != nil {
		return 0, err
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		// Check if directory name is a PID (numeric)
		pid, err := strconv.Atoi(dir.Name())
		if err != nil {
			continue
		}

		// Read cgroup file to check for container ID
		cgroupPath := fmt.Sprintf("%s/%d/cgroup", procBasePath, pid)
		cgroupFile, err := os.Open(cgroupPath)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(cgroupFile)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, containerID) {
				cgroupFile.Close()
				return pid, nil
			}
		}
		cgroupFile.Close()
	}

	return -1, fmt.Errorf("PID not found for container %s", containerID)
}
