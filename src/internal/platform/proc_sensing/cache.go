package proc_sensing

import (
	"sync"
	"time"
)

var (
	cachedProcesses []ProcessInfo
	cacheTimestamp  time.Time
	cacheMu         sync.RWMutex
	CacheTTL        = 500 * time.Millisecond
)

func GetAllProcessesCached() ([]ProcessInfo, error) {
	cacheMu.RLock()
	if time.Since(cacheTimestamp) < CacheTTL && cachedProcesses != nil {
		result := cachedProcesses
		cacheMu.RUnlock()
		return result, nil
	}
	cacheMu.RUnlock()

	cacheMu.Lock()
	defer cacheMu.Unlock()

	if time.Since(cacheTimestamp) < CacheTTL && cachedProcesses != nil {
		return cachedProcesses, nil
	}

	procs, err := GetAllProcesses()
	if err == nil {
		cachedProcesses = procs
		cacheTimestamp = time.Now()
	}
	return procs, err
}

func InvalidateProcessCache() {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cachedProcesses = nil
}
