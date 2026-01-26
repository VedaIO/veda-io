package app

import (
	"slices"
	"strings"
	"time"

	"wails-app/internal/blocklist/app"
	"wails-app/internal/data/logger"

	"github.com/shirou/gopsutil/v3/process"
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
			procs, err := process.Processes()
			if err != nil {
				appLogger.Printf("Failed to get processes: %v", err)
				continue
			}
			for _, p := range procs {
				name, _ := p.Name()
				if name == "" {
					continue
				}

				if slices.Contains(list, strings.ToLower(name)) {
					err := p.Kill()
					if err != nil {
						appLogger.Printf("failed to kill %s (pid %d): %v", name, p.Pid, err)
					} else {
						appLogger.Printf("killed blocked process %s (pid %d)", name, p.Pid)
					}
				}
			}
		}
	}()
}
