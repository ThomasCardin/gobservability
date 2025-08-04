package main

import (
	"fmt"

	"github.com/ThomasCardin/peek/pkg/kubernetes"
)

func main() {
	client, err := kubernetes.Client()
	if err != nil {
		fmt.Printf("error: %s \n", err.Error())
	}

	nodes, err := kubernetes.GetNodes(client)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}

	for _, n := range nodes.Items {
		pods, err := kubernetes.GetPods(client, n.Name)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			continue
		}

		for _, pod := range pods.Items {
			fmt.Printf("\nGetting container info for pod: %s\n", pod.Name)
			containerIDs, err := kubernetes.GetContainerID(client, pod.Name, pod.Namespace)
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				continue
			}

			for _, containerID := range containerIDs {
				pid, err := kubernetes.GetPID(containerID)
				if err != nil {
					fmt.Printf("%s\n", err.Error())
					continue
				}
				fmt.Printf("Success: Container %s has PID %d\n", containerID, pid)
			}
		}
	}
}
