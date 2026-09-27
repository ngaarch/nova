//go:build !windows

package sysinfo

import (
	"syscall"
)

// DiskSpace holds storage capacity statistics.
type DiskSpace struct {
	TotalBytes uint64  `json:"total_bytes"`
	FreeBytes  uint64  `json:"free_bytes"`
	UsedBytes  uint64  `json:"used_bytes"`
	UsedPct    float64 `json:"used_pct"`
	Available  bool    `json:"available"`
}

func getDiskSpace(path string) DiskSpace {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return DiskSpace{Available: false}
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free

	var pct float64
	if total > 0 {
		pct = float64(used) / float64(total) * 100.0
	}

	return DiskSpace{
		TotalBytes: total,
		FreeBytes:  free,
		UsedBytes:  used,
		UsedPct:    pct,
		Available:  true,
	}
}
