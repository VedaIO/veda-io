//go:build !windows

package proc_sensing

import (
	"github.com/shirou/gopsutil/v3/process"
)

// GetAllProcesses provides a pure Go implementation for Unix platforms using gopsutil.
func GetAllProcesses() ([]ProcessInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	results := make([]ProcessInfo, 0, len(procs))
	for _, p := range procs {
		name, _ := p.Name()
		exe, _ := p.Exe()
		// On Unix, gopsutil's CreateTime provides ms precision which is usually
		// enough to distinguish between recycled PIDs.
		createTime, _ := p.CreateTime()

		results = append(results, ProcessInfo{
			PID:           uint32(p.Pid),
			ParentPID:     0,                            // Can be resolved with p.Parent() if higher precision needed
			StartTimeNano: uint64(createTime) * 1000000, // ms to ns
			Name:          name,
			ExePath:       exe,
		})
	}

	return results, nil
}
