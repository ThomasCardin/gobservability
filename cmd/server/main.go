package server

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func clearScreen() {
	// Séquences d'échappement ANSI pour clear + cursor en haut
	fmt.Print("\033[2J\033[H")
}

func readCPUStats() string {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return fmt.Sprintf("Erreur: %v", err)
	}

	lines := strings.Split(string(data), "\n")
	return lines[0]
}

func main() {
	for {
		clearScreen()
		
		now := time.Now().Format("15:04:05")
		cpuStats := readCPUStats()
		
		fmt.Printf("Peek - CPU Monitor - %s\n\n", now)
		fmt.Printf("Stats CPU: %s\n", cpuStats)
		
		time.Sleep(5 * time.Second)
	}
}