package kubernetes

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func Client() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("error: failed to create Kubernetes config: %v", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error: failed to create Kubernetes client: %v", err)
	}

	return clientset, nil
}

func GetNodes(client *kubernetes.Clientset) (*v1.NodeList, error) {
	nodes, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error: failed to get nodes: %v", err)
	}

	return nodes, nil
}

func GetPods(client *kubernetes.Clientset, nodeName string) (*v1.PodList, error) {
	pods, err := client.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName,
	})
	if err != nil {
		return nil, fmt.Errorf("error: failed to get pods for node %s: %v", nodeName, err)
	}

	return pods, nil
}

func GetContainerID(client *kubernetes.Clientset, podName, namespace string) ([]string, error) {
	pod, err := client.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error: failed to get pod %s in namespace %s: %v", podName, namespace, err)
	}

	var containerIDs []string
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.ContainerID != "" {
			containerID := strings.TrimPrefix(containerStatus.ContainerID, "containerd://")
			containerID = strings.TrimPrefix(containerID, "docker://")
			containerIDs = append(containerIDs, containerID)
		}
	}

	return containerIDs, nil
}

func GetPID(containerID string) (int, error) {
	procDir := "/proc"

	err := filepath.WalkDir(procDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() && strings.HasPrefix(d.Name(), "/proc/") {
			return nil
		}

		if filepath.Base(filepath.Dir(path)) != "proc" {
			return nil
		}

		if _, err := strconv.Atoi(d.Name()); err != nil {
			return nil
		}

		cgroupPath := filepath.Join(path, "cgroup")
		content, err := os.ReadFile(cgroupPath)
		if err != nil {
			return nil
		}

		if strings.Contains(string(content), containerID) {
			pid, err := strconv.Atoi(d.Name())
			if err != nil {
				return nil
			}
			fmt.Printf("Container ID: %s, PID: %d\n", containerID, pid)
			return fmt.Errorf("found pid: %d", pid)
		}

		return nil
	})

	if err != nil && strings.Contains(err.Error(), "found pid:") {
		pidStr := strings.Split(err.Error(), ": ")[1]
		pid, _ := strconv.Atoi(pidStr)
		return pid, nil
	}

	return 0, fmt.Errorf("error: PID not found for container %s", containerID)
}
