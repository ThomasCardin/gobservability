package shared

import "os"

func GetProcBasePath(devMode string) string {
	if isDev := os.Getenv(devMode); isDev == "true" {
		return "/proc"
	}
	return "/host/proc"
}

// func getSysBasePath() string {
// 	// Check if running in dev mode
// 	if isDev := os.Getenv("AGENT_DEV_MODE"); isDev == "true" {
// 		return "/sys"
// 	}
// 	return "/host/sys"
// }
