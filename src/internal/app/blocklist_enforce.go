package app

import (
	"slices"
	"strings"
	"time"

	"os"
	"src/internal/blocklist/app"
	"src/internal/data/logger"
	"src/internal/platform/proc_sensing"
)

const blocklistEnforceInterval = 2 * time.Second

// StartBlocklistEnforcer starts a goroutine that kills processes listed in the blocklist.
func StartBlocklistEnforcer(appLogger logger.Logger) {
	go func() {
		killTick := time.NewTicker(blocklistEnforceInterval)
		defer killTick.Stop()
		for range killTick.C {
			list, err := app.LoadAppBlocklist()
			if err != nil {
				appLogger.Printf("failed to fetch blocklist: %v", err)
				continue
			}
			if len(list) == 0 {
				continue
			}
			procs, err := proc_sensing.GetAllProcesses()
			if err != nil {
				appLogger.Printf("Failed to get processes: %v", err)
				continue
			}
			for _, p := range procs {
				name := p.Name
				if name == "" {
					continue
				}

				if slices.Contains(list, strings.ToLower(name)) {
					osProc, err := os.FindProcess(int(p.PID))
					if err == nil {
						if err := osProc.Kill(); err != nil {
							appLogger.Printf("failed to kill %s (pid %d): %v", name, p.PID, err)
						} else {
							appLogger.Printf("killed blocked process %s (pid %d)", name, p.PID)
						}
					}
				}
			}

		}
	}()
}
